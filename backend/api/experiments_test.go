package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupExpRouter(t *testing.T) (*gin.Engine, int64) {
	t.Helper()
	db := setupTestDB(t)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	v1 := r.Group("/api/v1")
	v1.GET("/experiments", ListExperiments(db))
	v1.GET("/experiments/:id", GetExperiment(db))
	v1.POST("/experiments", CreateExperiment(db))
	v1.PUT("/experiments/:id", UpdateExperiment(db))

	// Insert a dataset to reference
	var datasetID int64
	err := db.QueryRow(`INSERT INTO datasets (name) VALUES ('test-dataset') RETURNING id`).Scan(&datasetID)
	if err != nil {
		t.Fatalf("insert dataset: %v", err)
	}
	return r, datasetID
}

func postExperiment(t *testing.T, r *gin.Engine, body map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/experiments", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestCreateExperiment_HappyPath(t *testing.T) {
	r, datasetID := setupExpRouter(t)

	rec := postExperiment(t, r, map[string]any{
		"name":         "exp-1",
		"dataset_id":   datasetID,
		"prompts":      []string{"system prompt A", "system prompt B"},
		"judge_prompt": "you are a judge",
	})

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		ID          int64    `json:"id"`
		Name        string   `json:"name"`
		Status      string   `json:"status"`
		Prompts     []string `json:"prompts"`
		JudgePrompt string   `json:"judge_prompt"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Name != "exp-1" {
		t.Errorf("expected name 'exp-1', got %q", resp.Name)
	}
	if resp.Status != "ready" {
		t.Errorf("expected status 'ready', got %q", resp.Status)
	}
	if len(resp.Prompts) != 2 {
		t.Errorf("expected 2 prompts, got %d", len(resp.Prompts))
	}
	if resp.JudgePrompt != "you are a judge" {
		t.Errorf("unexpected judge_prompt: %q", resp.JudgePrompt)
	}
}

func TestCreateExperiment_MissingName(t *testing.T) {
	r, datasetID := setupExpRouter(t)

	rec := postExperiment(t, r, map[string]any{
		"name":         "",
		"dataset_id":   datasetID,
		"prompts":      []string{"p"},
		"judge_prompt": "j",
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetExperiment_HappyPath(t *testing.T) {
	r, datasetID := setupExpRouter(t)

	created := postExperiment(t, r, map[string]any{
		"name":         "exp-get",
		"dataset_id":   datasetID,
		"prompts":      []string{"prompt X"},
		"judge_prompt": "judge Y",
	})
	if created.Code != http.StatusCreated {
		t.Fatalf("create failed: %d %s", created.Code, created.Body.String())
	}
	var c struct{ ID int64 `json:"id"` }
	json.NewDecoder(created.Body).Decode(&c)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/experiments/%d", c.ID), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Prompts     []string `json:"prompts"`
		JudgePrompt string   `json:"judge_prompt"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp.Prompts) != 1 || resp.Prompts[0] != "prompt X" {
		t.Errorf("unexpected prompts: %v", resp.Prompts)
	}
	if resp.JudgePrompt != "judge Y" {
		t.Errorf("unexpected judge_prompt: %q", resp.JudgePrompt)
	}
}

func TestGetExperiment_NotFound(t *testing.T) {
	r, _ := setupExpRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/experiments/99999", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestUpdateExperiment_StartHappyPath(t *testing.T) {
	r, datasetID := setupExpRouter(t)

	created := postExperiment(t, r, map[string]any{
		"name": "exp-run", "dataset_id": datasetID,
		"prompts": []string{"p"}, "judge_prompt": "j",
	})
	var c struct{ ID int64 `json:"id"` }
	json.NewDecoder(created.Body).Decode(&c)

	body, _ := json.Marshal(map[string]string{"status": "in progress"})
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/experiments/%d", c.ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct{ Status string `json:"status"` }
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Status != "in progress" {
		t.Errorf("expected status 'in progress', got %q", resp.Status)
	}
}

func TestUpdateExperiment_AlreadyStarted(t *testing.T) {
	r, datasetID := setupExpRouter(t)

	created := postExperiment(t, r, map[string]any{
		"name": "exp-double", "dataset_id": datasetID,
		"prompts": []string{"p"}, "judge_prompt": "j",
	})
	var c struct{ ID int64 `json:"id"` }
	json.NewDecoder(created.Body).Decode(&c)

	body, _ := json.Marshal(map[string]string{"status": "in progress"})
	put := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/experiments/%d", c.ID), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		return rec
	}
	put() // first call — succeeds
	rec := put()
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateExperiment_InvalidStatus(t *testing.T) {
	r, datasetID := setupExpRouter(t)

	created := postExperiment(t, r, map[string]any{
		"name": "exp-bad", "dataset_id": datasetID,
		"prompts": []string{"p"}, "judge_prompt": "j",
	})
	var c struct{ ID int64 `json:"id"` }
	json.NewDecoder(created.Body).Decode(&c)

	body, _ := json.Marshal(map[string]string{"status": "done"})
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/experiments/%d", c.ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
