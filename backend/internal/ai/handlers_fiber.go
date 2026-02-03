package ai

import (
	"database/sql"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/webide/ide/backend/internal/db"
	"github.com/webide/ide/backend/internal/models"
)

func RegisterRoutes(router fiber.Router) {

	aiTasks := router.Group("/projects/:id/ai/tasks")
	aiTasks.Post("", HandleCreateTaskFiber)

	aiJobs := router.Group("/projects/:id/jobs")
	aiJobs.Get("", HandleListJobsFiber)

	jobRoutes := router.Group("/jobs/:jobId")
	jobRoutes.Get("", HandleGetJobFiber)

	aiChangeSets := router.Group("/projects/:id/changesets")
	aiChangeSets.Get("", HandleListChangeSetsFiber)

	csRoutes := router.Group("/changesets/:csId")
	csRoutes.Get("", HandleGetChangeSetFiber)
	csRoutes.Post("/apply", HandleApplyChangeSetFiber)
	csRoutes.Post("/revert", HandleRevertChangeSetFiber)
	csRoutes.Post("/approve", HandleApproveChangeSetFiber)
	csRoutes.Post("/request-changes", HandleRequestChangesFiber)
	csRoutes.Post("/threads", HandleCreateThreadFiber)

	threadRoutes := router.Group("/threads/:threadId")
	threadRoutes.Post("/comments", HandleAddCommentFiber)
	threadRoutes.Post("/resolve", HandleResolveThreadFiber)

	sendToAI := router.Group("/projects/:id/ai/send-to-ai")
	sendToAI.Post("", HandleSendToAIFiber)
}

func HandleCreateTaskFiber(c *fiber.Ctx) error {
	ctx := c.Context()
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project_id"})
	}

	var req AITaskRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.Prompt == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "prompt is required"})
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create job"})
	}

	go processAITask(dbJob.ID, projectID, req)

	return c.JSON(fiber.Map{
		"job_id": dbJob.ID,
		"status": "queued",
	})
}

func HandleListJobsFiber(c *fiber.Ctx) error {
	ctx := c.Context()
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project_id"})
	}

	rows, err := db.Query(ctx, "SELECT id, project_id, type, status, COALESCE(payload_json, ''), COALESCE(result_json, ''), COALESCE(error_text, ''), created_at, started_at, finished_at FROM jobs WHERE project_id = $1 ORDER BY created_at DESC", projectID.String())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to query jobs"})
	}
	defer rows.Close()

	var jobs []models.Job
	for rows.Next() {
		var job models.Job
		var payloadJSON, resultJSON, errorText sql.NullString
		err := rows.Scan(&job.ID, &job.ProjectID, &job.Type, &job.Status, &payloadJSON, &resultJSON, &errorText, &job.CreatedAt, &job.StartedAt, &job.FinishedAt)
		if err != nil {
			continue
		}
		job.PayloadJSON = payloadJSON.String
		job.ResultJSON = resultJSON.String
		job.ErrorText = errorText.String
		jobs = append(jobs, job)
	}

	return c.JSON(jobs)
}

func HandleGetJobFiber(c *fiber.Ctx) error {
	ctx := c.Context()
	jobID, err := uuid.Parse(c.Params("jobId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid job_id"})
	}

	rows, err := db.Query(ctx, "SELECT id, project_id, type, status, COALESCE(payload_json, ''), COALESCE(result_json, ''), COALESCE(error_text, ''), created_at, started_at, finished_at FROM jobs WHERE id = $1", jobID.String())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to query job"})
	}
	defer rows.Close()

	if !rows.Next() {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "job not found"})
	}

	var job models.Job
	var payloadJSON, resultJSON, errorText sql.NullString
	err = rows.Scan(&job.ID, &job.ProjectID, &job.Type, &job.Status, &payloadJSON, &resultJSON, &errorText, &job.CreatedAt, &job.StartedAt, &job.FinishedAt)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to scan job"})
	}
	job.PayloadJSON = payloadJSON.String
	job.ResultJSON = resultJSON.String
	job.ErrorText = errorText.String

	return c.JSON(job)
}

func HandleListChangeSetsFiber(c *fiber.Ctx) error {
	ctx := c.Context()
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project_id"})
	}

	rows, err := db.Query(ctx, "SELECT id, project_id, COALESCE(job_id, ''), title, base_ref, COALESCE(target_ref, ''), apply_mode, status, COALESCE(summary_text, ''), created_at, updated_at FROM changesets WHERE project_id = $1 ORDER BY created_at DESC", projectID.String())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to query changesets"})
	}
	defer rows.Close()

	var changesets []models.ChangeSet
	for rows.Next() {
		var cs models.ChangeSet
		var jobID, targetRef, summaryText sql.NullString
		err := rows.Scan(&cs.ID, &cs.ProjectID, &jobID, &cs.Title, &cs.BaseRef, &targetRef, &cs.ApplyMode, &cs.Status, &summaryText, &cs.CreatedAt, &cs.UpdatedAt)
		if err != nil {
			continue
		}
		if jobID.Valid && jobID.String != "" {
			id := uuid.MustParse(jobID.String)
			cs.JobID = &id
		}
		if targetRef.Valid {
			cs.TargetRef = &targetRef.String
		}
		cs.SummaryText = summaryText.String
		changesets = append(changesets, cs)
	}

	return c.JSON(changesets)
}

func HandleGetChangeSetFiber(c *fiber.Ctx) error {
	ctx := c.Context()
	csID, err := uuid.Parse(c.Params("csId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid changeset_id"})
	}

	rows, err := db.Query(ctx, "SELECT id, project_id, COALESCE(job_id, ''), title, base_ref, COALESCE(target_ref, ''), apply_mode, status, COALESCE(summary_text, ''), created_at, updated_at FROM changesets WHERE id = $1", csID.String())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to query changeset"})
	}
	defer rows.Close()

	if !rows.Next() {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "changeset not found"})
	}

	var cs models.ChangeSet
	var jobID, targetRef, summaryText sql.NullString
	err = rows.Scan(&cs.ID, &cs.ProjectID, &jobID, &cs.Title, &cs.BaseRef, &targetRef, &cs.ApplyMode, &cs.Status, &summaryText, &cs.CreatedAt, &cs.UpdatedAt)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to scan changeset"})
	}
	if jobID.Valid && jobID.String != "" {
		id := uuid.MustParse(jobID.String)
		cs.JobID = &id
	}
	if targetRef.Valid {
		cs.TargetRef = &targetRef.String
	}
	cs.SummaryText = summaryText.String

	return c.JSON(fiber.Map{
		"changeset": cs,
	})
}

func HandleApplyChangeSetFiber(c *fiber.Ctx) error {
	ctx := c.Context()
	csID, err := uuid.Parse(c.Params("csId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid changeset_id"})
	}

	_, err = db.Exec(ctx, "UPDATE changesets SET status = 'applied', updated_at = $1 WHERE id = $2",
		time.Now(), csID.String())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update changeset"})
	}

	return c.SendStatus(fiber.StatusOK)
}

func HandleRevertChangeSetFiber(c *fiber.Ctx) error {
	ctx := c.Context()
	csID, err := uuid.Parse(c.Params("csId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid changeset_id"})
	}

	_, err = db.Exec(ctx, "UPDATE changesets SET status = 'reverted', updated_at = $1 WHERE id = $2",
		time.Now(), csID.String())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update changeset"})
	}

	return c.SendStatus(fiber.StatusOK)
}

func HandleApproveChangeSetFiber(c *fiber.Ctx) error {
	ctx := c.Context()
	csID, err := uuid.Parse(c.Params("csId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid changeset_id"})
	}

	_, err = db.Exec(ctx, "UPDATE changesets SET status = 'approved', updated_at = $1 WHERE id = $2",
		time.Now(), csID.String())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update changeset"})
	}

	return c.SendStatus(fiber.StatusOK)
}

func HandleRequestChangesFiber(c *fiber.Ctx) error {
	ctx := c.Context()
	csID, err := uuid.Parse(c.Params("csId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid changeset_id"})
	}

	var req struct {
		Comment string `json:"comment"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	now := time.Now()
	threadID := uuid.New()

	_, err = db.Exec(ctx,
		"INSERT INTO review_threads (id, changeset_id, file_path, anchor_json, status, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		threadID, csID, "", "{}", "open", now)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create thread"})
	}

	_, err = db.Exec(ctx,
		"INSERT INTO review_comments (id, thread_id, author_user_id, body, created_at) VALUES ($1, $2, $3, $4, $5)",
		uuid.New(), threadID, "", req.Comment, now)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create comment"})
	}

	_, err = db.Exec(ctx, "UPDATE changesets SET status = 'changes_requested', updated_at = $1 WHERE id = $2",
		now, csID.String())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update changeset"})
	}

	return c.SendStatus(fiber.StatusOK)
}

func HandleCreateThreadFiber(c *fiber.Ctx) error {
	ctx := c.Context()
	csID, err := uuid.Parse(c.Params("csId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid changeset_id"})
	}

	var req struct {
		FilePath string `json:"filePath"`
		Anchor   string `json:"anchor"`
		Body     string `json:"body"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	now := time.Now()
	threadID := uuid.New()

	_, err = db.Exec(ctx,
		"INSERT INTO review_threads (id, changeset_id, file_path, anchor_json, status, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		threadID, csID, req.FilePath, req.Anchor, "open", now)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create thread"})
	}

	_, err = db.Exec(ctx,
		"INSERT INTO review_comments (id, thread_id, author_user_id, body, created_at) VALUES ($1, $2, $3, $4, $5)",
		uuid.New(), threadID, "", req.Body, now)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create comment"})
	}

	return c.JSON(fiber.Map{"thread_id": threadID})
}

func HandleAddCommentFiber(c *fiber.Ctx) error {
	ctx := c.Context()
	threadID, err := uuid.Parse(c.Params("threadId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid thread_id"})
	}

	var req struct {
		Body string `json:"body"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	_, err = db.Exec(ctx,
		"INSERT INTO review_comments (id, thread_id, author_user_id, body, created_at) VALUES ($1, $2, $3, $4, $5)",
		uuid.New(), threadID, "", req.Body, time.Now())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create comment"})
	}

	return c.SendStatus(fiber.StatusOK)
}

func HandleResolveThreadFiber(c *fiber.Ctx) error {
	ctx := c.Context()
	threadID, err := uuid.Parse(c.Params("threadId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid thread_id"})
	}

	_, err = db.Exec(ctx, "UPDATE review_threads SET status = 'resolved' WHERE id = $1", threadID.String())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to resolve thread"})
	}

	return c.SendStatus(fiber.StatusOK)
}

func HandleSendToAIFiber(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNotImplemented)
}
