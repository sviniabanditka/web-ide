package ai

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/webide/ide/backend/internal/config"
)

type UsageResponse struct {
	RemainingCredits float64 `json:"remaining_credits"`
	TotalCredits     float64 `json:"total_credits"`
	UsedCredits      float64 `json:"used_credits"`
	PercentUsed      float64 `json:"percent_used"`
	TotalCount       int     `json:"total_count"`
	UsageCount       int     `json:"usage_count"`
	ModelName        string  `json:"model_name"`
	StartTime        int64   `json:"start_time"`
	EndTime          int64   `json:"end_time"`
}

type ModelRemain struct {
	StartTime                 int64  `json:"start_time"`
	EndTime                   int64  `json:"end_time"`
	RemainsTime               int64  `json:"remains_time"`
	CurrentIntervalTotalCount int    `json:"current_interval_total_count"`
	CurrentIntervalUsageCount int    `json:"current_interval_usage_count"`
	ModelName                 string `json:"model_name"`
}

type MiniMaxUsageResponse struct {
	ModelRemains []ModelRemain `json:"model_remains"`
	BaseResp     struct {
		StatusCode int    `json:"status_code"`
		StatusMsg  string `json:"status_msg"`
	} `json:"base_resp"`
}

func RegisterUsageRoutes(router fiber.Router, cfg *config.Config) {
	log.Println("RegisterUsageRoutes: starting...")

	usage := router.Group("/ai/usage")
	log.Println("RegisterUsageRoutes: created group /ai/usage")

	usage.Get("", HandleGetUsage(cfg))

	log.Println("RegisterUsageRoutes: all routes registered")
}

func HandleGetUsage(cfg *config.Config) fiber.Handler {
	log.Printf("[HandleGetUsage] Starting")
	return func(c *fiber.Ctx) error {
		if cfg.MiniMaxAPIKey == "" {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "MiniMax API key not configured"})
		}

		url := "https://www.minimax.io/v1/api/openplatform/coding_plan/remains"
		req, err := http.NewRequestWithContext(c.Context(), "GET", url, nil)
		if err != nil {
			log.Printf("[HandleGetUsage] Failed to create request: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create request"})
		}

		req.Header.Set("Authorization", "Bearer "+cfg.MiniMaxAPIKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")
		req.Header.Set("Origin", "https://www.minimax.io")
		req.Header.Set("Referer", "https://www.minimax.io/")

		log.Printf("[HandleGetUsage] Request headers:")
		for k, v := range req.Header {
			log.Printf("[HandleGetUsage]   %s: %v", k, v)
		}

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("[HandleGetUsage] Request failed: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch usage"})
		}
		defer resp.Body.Close()

		log.Printf("[HandleGetUsage] Response status: %d", resp.StatusCode)
		log.Printf("[HandleGetUsage] Response headers:")
		for k, v := range resp.Header {
			log.Printf("[HandleGetUsage]   %s: %v", k, v)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("[HandleGetUsage] Failed to read response: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to read response"})
		}

		log.Printf("[HandleGetUsage] Full response body:")
		log.Printf("%s", string(body))

		if resp.StatusCode != 200 {
			log.Printf("[HandleGetUsage] API error status %d", resp.StatusCode)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "api error", "details": string(body)})
		}

		var apiResp MiniMaxUsageResponse
		if err := json.Unmarshal(body, &apiResp); err != nil {
			log.Printf("[HandleGetUsage] Failed to parse response: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to parse response"})
		}

		log.Printf("[HandleGetUsage] Base resp: status_code=%d, status_msg=%s",
			apiResp.BaseResp.StatusCode, apiResp.BaseResp.StatusMsg)

		if len(apiResp.ModelRemains) == 0 {
			log.Printf("[HandleGetUsage] No model remains found")
			return c.JSON(UsageResponse{
				RemainingCredits: 0,
				TotalCredits:     0,
				UsedCredits:      0,
				PercentUsed:      0,
			})
		}

		model := apiResp.ModelRemains[0]
		totalCount := model.CurrentIntervalTotalCount
		remainingCount := model.CurrentIntervalUsageCount
		usedCount := totalCount - remainingCount

		percentUsed := 0.0
		if totalCount > 0 {
			percentUsed = float64(usedCount) / float64(totalCount) * 100
		}

		log.Printf("[HandleGetUsage] Parsed: model=%s, total=%d, used=%d, remaining=%d, percent=%.2f%%",
			model.ModelName, totalCount, usedCount, remainingCount, percentUsed)

		return c.JSON(UsageResponse{
			RemainingCredits: float64(remainingCount),
			TotalCredits:     float64(totalCount),
			UsedCredits:      float64(usedCount),
			PercentUsed:      percentUsed,
			TotalCount:       totalCount,
			UsageCount:       usedCount,
			ModelName:        model.ModelName,
			StartTime:        model.StartTime,
			EndTime:          model.EndTime,
		})
	}
}
