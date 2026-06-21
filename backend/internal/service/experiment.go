package service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"fsa-llm-experiments/backend/dal"
)

// ChatFunc sends a system prompt and user message to an LLM and returns the response.
type ChatFunc func(ctx context.Context, systemPrompt, userMsg string) (string, error)

func ollamaBaseURL() string {
	if u := os.Getenv("OLLAMA_URL"); u != "" {
		return u
	}
	return "http://host.docker.internal:11434"
}

func ollamaModel() string {
	if m := os.Getenv("OLLAMA_MODEL"); m != "" {
		return m
	}
	return "qwen3:8b"
}

type ollamaRequest struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaResponse struct {
	Message message `json:"message"`
}

// OllamaChat is the production ChatFunc that calls the local Ollama API.
func OllamaChat(ctx context.Context, systemPrompt, userMsg string) (string, error) {
	client := &http.Client{}
	payload, _ := json.Marshal(ollamaRequest{
		Model: ollamaModel(),
		Messages: []message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMsg},
		},
		Stream: false,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ollamaBaseURL()+"/api/chat", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	var out ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.Message.Content, nil
}

// RunExperiment runs all prompt variants against all dataset queries, judges each
// response, computes a pass-rate score (0–100), then calls dal.FinishExperiment.
// Intended to be called in a background goroutine with context.Background().
// Pass nil for chat to use the default OllamaChat implementation.
func RunExperiment(ctx context.Context, db *sql.DB, experimentID int64, chat ChatFunc) error {
	if chat == nil {
		chat = OllamaChat
	}

	log := slog.With("experiment_id", experimentID)
	log.Info("experiment started")

	exp, err := dal.GetExperiment(ctx, db, experimentID)
	if err != nil {
		log.Error("failed to load experiment", "error", err)
		dal.FinishExperiment(ctx, db, experimentID, 0, true)
		return err
	}

	queries, err := dal.ListQueries(ctx, db, exp.DatasetID)
	if err != nil {
		log.Error("failed to load queries", "dataset_id", exp.DatasetID, "error", err)
		dal.FinishExperiment(ctx, db, experimentID, 0, true)
		return err
	}

	total := len(exp.Prompts) * len(queries)
	var passes int
	completed := 0

	log.Info("experiment running", "variants", len(exp.Prompts), "queries", len(queries), "total_pairs", total)

	for variantIdx, prompt := range exp.Prompts {
		for queryIdx, query := range queries {
			completed++
			log.Info("calling model",
				"variant", variantIdx+1,
				"query", queryIdx+1,
				"progress", fmt.Sprintf("%d/%d", completed, total),
				"system_prompt", prompt,
				"user_msg", query,
			)

			modelResponse, err := chat(ctx, prompt, query)
			if err != nil {
				log.Error("model call failed", "variant", variantIdx+1, "query", queryIdx+1, "error", err)
				dal.FinishExperiment(ctx, db, experimentID, 0, true)
				return err
			}

			judgeInput := fmt.Sprintf("Query: %s\n\nResponse: %s", query, modelResponse)
			log.Info("calling judge",
				"variant", variantIdx+1,
				"query", queryIdx+1,
				"system_prompt", exp.JudgePrompt,
				"user_msg", judgeInput,
			)

			verdict, err := chat(ctx, exp.JudgePrompt, judgeInput)
			if err != nil {
				log.Error("judge call failed", "variant", variantIdx+1, "query", queryIdx+1, "error", err)
				dal.FinishExperiment(ctx, db, experimentID, 0, true)
				return err
			}

			pass := strings.Contains(strings.ToUpper(verdict), "PASS")
			if pass {
				passes++
			}
			log.Info("query judged",
				"variant", variantIdx+1,
				"query", queryIdx+1,
				"result", map[bool]string{true: "PASS", false: "FAIL"}[pass],
				"passes_so_far", passes,
			)
		}
	}

	var score float64
	if total > 0 {
		score = float64(passes) / float64(total) * 100
	}

	log.Info("experiment finished", "passes", passes, "total", total, "score", fmt.Sprintf("%.1f%%", score))
	return dal.FinishExperiment(ctx, db, experimentID, score, false)
}
