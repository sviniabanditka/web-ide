package ai

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/webide/ide/backend/internal/ai/provider"
	"github.com/webide/ide/backend/internal/db"
	"github.com/webide/ide/backend/internal/git"
	"github.com/webide/ide/backend/internal/models"
	"github.com/webide/ide/backend/internal/projects"
)

type AITaskRequest struct {
	Prompt        string                 `json:"prompt"`
	Context       map[string]interface{} `json:"context,omitempty"`
	Mode          string                 `json:"mode,omitempty"`
	MaxFiles      int                    `json:"max_files,omitempty"`
	DisallowPaths []string               `json:"disallow_paths,omitempty"`
}

type LLMResult struct {
	Summary     string              `json:"summary"`
	Plan        string              `json:"plan"`
	UnifiedDiff string              `json:"unified_diff"`
	Notes       string              `json:"notes"`
	Usage       provider.TokenUsage `json:"usage"`
}

func HandleCreateTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	projectID, _ := uuid.Parse(vars["id"])
	if projectID == uuid.Nil {
		http.Error(w, "invalid project_id", http.StatusBadRequest)
		return
	}

	var req AITaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Prompt == "" {
		http.Error(w, "prompt is required", http.StatusBadRequest)
		return
	}

	if req.Mode == "" {
		req.Mode = "patch_to_working_tree"
	}
	if req.MaxFiles == 0 {
		req.MaxFiles = 20
	}

	dbJob := &models.Job{
		ID:          uuid.New(),
		ProjectID:   projectID,
		Type:        "ai_task",
		Status:      "queued",
		PayloadJSON: mustMarshal(req),
		CreatedAt:   time.Now(),
	}

	if err := db.Insert(ctx, "jobs", dbJob); err != nil {
		http.Error(w, "failed to create job", http.StatusInternalServerError)
		return
	}

	go processAITask(dbJob.ID, projectID, req)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"job_id": dbJob.ID,
		"status": "queued",
	})
}

func processAITask(jobID, projectID uuid.UUID, req AITaskRequest) {
	log.Printf("processAITask: started for jobID=%s, projectID=%s", jobID.String(), projectID.String())
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	job, err := getJobByID(ctx, jobID)
	if err != nil || job == nil {
		log.Printf("processAITask: job not found or error: %v", err)
		return
	}

	now := time.Now()
	job.Status = "running"
	job.StartedAt = &now
	db.Update(ctx, "jobs", job)
	BroadcastJobUpdate(projectID.String(), jobID.String(), "running", "", nil)

	project, err := projects.GetProject(projectID)
	if err != nil {
		updateJobError(ctx, jobID, "project not found")
		BroadcastJobUpdate(projectID.String(), jobID.String(), "failed", "project not found", nil)
		return
	}

	messages, llmResp, err := callLLM(ctx, project.RootPath, req)
	if err != nil {
		updateJobError(ctx, jobID, err.Error())
		BroadcastJobUpdate(projectID.String(), jobID.String(), "failed", err.Error(), nil)
		return
	}

	aiContext := map[string]interface{}{
		"prompt":      req.Prompt,
		"messages":    messages,
		"model_usage": llmResp.Usage,
	}
	aiContextJSON := mustMarshal(aiContext)

	result := map[string]interface{}{
		"summary": llmResp.Summary,
		"plan":    llmResp.Plan,
		"notes":   llmResp.Notes,
		"diff":    llmResp.UnifiedDiff,
	}

	if err := saveJobResult(ctx, jobID, result, aiContextJSON); err != nil {
		updateJobError(ctx, jobID, "failed to save result")
		BroadcastJobUpdate(projectID.String(), jobID.String(), "failed", "failed to save result", nil)
		return
	}

	if req.Mode == "patch_to_working_tree" && llmResp.UnifiedDiff != "" {
		if err := applyPatch(project.RootPath, llmResp.UnifiedDiff); err != nil {
			updateJobError(ctx, jobID, "failed to apply patch: "+err.Error())
			BroadcastJobUpdate(projectID.String(), jobID.String(), "failed", "failed to apply patch", nil)
			return
		}

		if err := createChangesetFromDiff(ctx, projectID, jobID, llmResp); err != nil {
			updateJobError(ctx, jobID, "failed to create changeset: "+err.Error())
			BroadcastJobUpdate(projectID.String(), jobID.String(), "failed", "failed to create changeset", nil)
			return
		}
	}

	finishedAt := time.Now()
	job.Status = "succeeded"
	job.FinishedAt = &finishedAt
	db.Update(ctx, "jobs", job)
	BroadcastJobUpdate(projectID.String(), jobID.String(), "succeeded", "", result)
}

func callLLM(ctx context.Context, projectRoot string, req AITaskRequest) ([]provider.Message, *LLMResult, error) {
	systemMsg := buildSystemPrompt(req)
	messages := []provider.Message{
		{Role: "system", Content: systemMsg},
	}

	if selectedText, ok := req.Context["selectedText"].(string); ok && selectedText != "" {
		messages = append(messages, provider.Message{
			Role:    "user",
			Content: "Selected code:\n```\n" + selectedText + "\n```",
		})
	}

	messages = append(messages, provider.Message{
		Role:    "user",
		Content: req.Prompt,
	})

	cfg := provider.Config{
		URL:       "",
		APIKey:    "",
		Model:     "MiniMax-ABAB",
		MaxTokens: 8192,
	}

	resp, err := provider.Complete(ctx, provider.ProviderMiniMax, messages, cfg)
	if err != nil {
		return nil, nil, err
	}

	llmResp := parseLLMResponse(resp.Content)
	llmResp.Usage = resp.Usage
	return messages, llmResp, nil
}

func buildSystemPrompt(req AITaskRequest) string {
	return `You are an expert software developer AI assistant. Generate unified diffs for code changes.

Output format (JSON):
{
  "summary": "Brief description of changes",
  "plan": "Step-by-step plan",
  "unified_diff": "完整的unified diff格式的代码修改",
  "notes": "Any important notes or caveats"
}

Rules:
1. 只修改明确要求的代码，不要添加无关的修改
2. Maintain existing code style
3. 只修改与任务相关的文件
4. Use standard unified diff format
5. Include only files that need to change
6. 如果无法完成任务，返回空的unified_diff但提供summary和plan说明情况`
}

func parseLLMResponse(content string) *LLMResult {
	resp := &LLMResult{}

	jsonStart := strings.Index(content, "{")
	jsonEnd := strings.LastIndex(content, "}")
	if jsonStart == -1 || jsonEnd == -1 {
		resp.Summary = content
		return resp
	}

	jsonStr := content[jsonStart : jsonEnd+1]
	if err := json.Unmarshal([]byte(jsonStr), resp); err != nil {
		resp.Summary = content
	}

	resp.UnifiedDiff = extractDiff(resp.UnifiedDiff)
	return resp
}

func extractDiff(text string) string {
	if strings.HasPrefix(text, "```diff") {
		lines := strings.Split(text, "\n")
		if len(lines) >= 2 {
			return strings.Join(lines[1:len(lines)-1], "\n")
		}
	}
	if strings.HasPrefix(text, "diff --git") {
		return text
	}
	return text
}

func applyPatch(projectRoot, diff string) error {
	if diff == "" || !strings.HasPrefix(diff, "diff --git") {
		return nil
	}
	return git.ApplyPatch(projectRoot, diff)
}

func createChangesetFromDiff(ctx context.Context, projectID, jobID uuid.UUID, llmResp *LLMResult) error {
	headRef, _ := git.GetHeadCommit("")

	cs := &models.ChangeSet{
		ID:          uuid.New(),
		ProjectID:   projectID,
		JobID:       &jobID,
		Title:       llmResp.Summary,
		BaseRef:     headRef,
		ApplyMode:   "working_tree",
		Status:      "needs_review",
		SummaryText: llmResp.Plan,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := db.Insert(ctx, "changesets", cs); err != nil {
		return err
	}

	BroadcastChangeSetCreated(projectID.String(), cs.ID.String(), cs.Title, cs.Status, cs.SummaryText)

	files := parseDiffFiles(llmResp.UnifiedDiff)
	for _, f := range files {
		_, err := db.Exec(ctx,
			"INSERT INTO changeset_files (id, changeset_id, path) VALUES ($1, $2, $3)",
			uuid.New(), cs.ID, f)
		if err != nil {
			return err
		}
	}

	return nil
}

func parseDiffFiles(diff string) []string {
	var files []string
	re := regexp.MustCompile(`diff --git a/(\S+) b/(\S+)`)
	matches := re.FindAllStringSubmatch(diff, -1)
	for _, m := range matches {
		if len(m) >= 3 {
			files = append(files, m[2])
		}
	}
	return files
}

func getJobByID(ctx context.Context, id uuid.UUID) (*models.Job, error) {
	var job models.Job
	err := db.Get(ctx, &job, "SELECT * FROM jobs WHERE id = $1", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func updateJobStatus(ctx context.Context, id uuid.UUID, status string, startedAt *time.Time) error {
	_, err := db.Exec(ctx,
		"UPDATE jobs SET status = $1, started_at = $2 WHERE id = $3",
		status, startedAt, id)
	return err
}

func updateJobError(ctx context.Context, id uuid.UUID, errorText string) error {
	now := time.Now()
	_, err := db.Exec(ctx,
		"UPDATE jobs SET status = $1, error_text = $2, finished_at = $3 WHERE id = $4",
		"failed", errorText, now, id)
	return err
}

func saveJobResult(ctx context.Context, id uuid.UUID, result map[string]interface{}, contextJSON string) error {
	resultJSON, _ := json.Marshal(result)
	_, err := db.Exec(ctx,
		"UPDATE jobs SET result_json = $1, finished_at = $2 WHERE id = $3",
		string(resultJSON), time.Now(), id)
	return err
}

func mustMarshal(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func ptr[T any](v T) *T {
	return &v
}
