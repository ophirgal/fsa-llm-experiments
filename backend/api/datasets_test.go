package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS datasets (
			id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE TABLE IF NOT EXISTS user_queries (
			id BIGSERIAL PRIMARY KEY,
			dataset_id BIGINT NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
			query TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	if err != nil {
		t.Fatalf("create tables: %v", err)
	}
	t.Cleanup(func() {
		db.Exec(`DELETE FROM user_queries`)
		db.Exec(`DELETE FROM datasets`)
	})
	return db
}

func setupRouter(db *sql.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	v1 := r.Group("/api/v1")
	v1.GET("/datasets", ListDatasets(db))
	v1.POST("/datasets", CreateDataset(db))
	return r
}

func multipartUpload(name, fileContent string) (*bytes.Buffer, string) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("name", name)
	fw, _ := w.CreateFormFile("file", "queries.txt")
	fmt.Fprint(fw, fileContent)
	w.Close()
	return &buf, w.FormDataContentType()
}

func TestCreateDataset_HappyPath(t *testing.T) {
	db := setupTestDB(t)
	r := setupRouter(db)

	body, ct := multipartUpload("my dataset", "query one\nquery two\nquery three\n")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/datasets", body)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		ID         int64  `json:"id"`
		Name       string `json:"name"`
		QueryCount int    `json:"query_count"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Name != "my dataset" {
		t.Errorf("expected name 'my dataset', got %q", resp.Name)
	}
	if resp.QueryCount != 3 {
		t.Errorf("expected query_count 3, got %d", resp.QueryCount)
	}
}

func TestCreateDataset_MissingFile(t *testing.T) {
	db := setupTestDB(t)
	r := setupRouter(db)

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("name", "no file")
	w.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/datasets", strings.NewReader(buf.String()))
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
