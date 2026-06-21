package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"fsa-llm-experiments/backend/dal"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}

	b := make([]byte, 8)
	rand.Read(b)
	schema := "test_" + hex.EncodeToString(b)

	base, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if _, err = base.Exec("CREATE SCHEMA " + schema); err != nil {
		base.Close()
		t.Fatalf("create schema: %v", err)
	}
	base.Close()

	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("parse dsn: %v", err)
	}
	q := u.Query()
	q.Set("options", fmt.Sprintf("-c search_path=%s", schema))
	u.RawQuery = q.Encode()

	db, err := sql.Open("pgx", u.String())
	if err != nil {
		t.Fatalf("open schema db: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
		drop, _ := sql.Open("pgx", dsn)
		drop.Exec("DROP SCHEMA " + schema + " CASCADE")
		drop.Close()
	})

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
		CREATE TYPE experiment_status AS ENUM ('ready', 'in progress', 'done', 'failed');
		CREATE TABLE IF NOT EXISTS experiments (
			id          BIGSERIAL PRIMARY KEY,
			name        TEXT               NOT NULL,
			dataset_id  BIGINT             NOT NULL REFERENCES datasets(id) ON DELETE RESTRICT,
			status      experiment_status  NOT NULL DEFAULT 'ready',
			total_score DOUBLE PRECISION,
			start_time  TIMESTAMPTZ,
			end_time    TIMESTAMPTZ,
			created_at  TIMESTAMPTZ        NOT NULL DEFAULT NOW()
		);
		CREATE TABLE IF NOT EXISTS experiment_prompts (
			id            BIGSERIAL PRIMARY KEY,
			experiment_id BIGINT      NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
			prompt        TEXT        NOT NULL,
			created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE TABLE IF NOT EXISTS experiment_judge_prompt (
			id            BIGSERIAL PRIMARY KEY,
			experiment_id BIGINT      NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
			prompt        TEXT        NOT NULL,
			created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	if err != nil {
		t.Fatalf("create tables: %v", err)
	}
	return db
}

// seedExperiment creates a dataset with queries, creates an experiment, starts it,
// and returns the experiment ID.
func seedExperiment(t *testing.T, db *sql.DB, queries []string) int64 {
	t.Helper()
	dataset, err := dal.CreateDataset(context.Background(), db, "test-dataset", queries)
	if err != nil {
		t.Fatalf("create dataset: %v", err)
	}
	exp, err := dal.CreateExperiment(context.Background(), db, "test-exp", dataset.ID,
		[]string{"system prompt A", "system prompt B"}, "you are a judge. reply PASS or FAIL")
	if err != nil {
		t.Fatalf("create experiment: %v", err)
	}
	if _, err := dal.StartExperiment(context.Background(), db, exp.ID); err != nil {
		t.Fatalf("start experiment: %v", err)
	}
	return exp.ID
}

func TestRunExperiment_AllPass(t *testing.T) {
	db := setupTestDB(t)
	id := seedExperiment(t, db, []string{"query one", "query two"})

	// Mock: model always returns "great answer", judge always returns "PASS"
	callN := 0
	mockChat := ChatFunc(func(ctx context.Context, system, user string) (string, error) {
		callN++
		if system == "you are a judge. reply PASS or FAIL" {
			return "PASS", nil
		}
		return "great answer", nil
	})

	if err := RunExperiment(context.Background(), db, id, mockChat); err != nil {
		t.Fatalf("RunExperiment: %v", err)
	}

	// 2 prompts × 2 queries = 4 model calls + 4 judge calls = 8 total
	if callN != 8 {
		t.Errorf("expected 8 chat calls, got %d", callN)
	}

	exp, err := dal.GetExperiment(context.Background(), db, id)
	if err != nil {
		t.Fatalf("get experiment: %v", err)
	}
	if exp.Status != "done" {
		t.Errorf("expected status 'done', got %q", exp.Status)
	}
	if exp.TotalScore == nil || *exp.TotalScore != 100 {
		t.Errorf("expected score 100, got %v", exp.TotalScore)
	}
}

func TestRunExperiment_AllFail(t *testing.T) {
	db := setupTestDB(t)
	id := seedExperiment(t, db, []string{"query one"})

	mockChat := ChatFunc(func(ctx context.Context, system, user string) (string, error) {
		if system == "you are a judge. reply PASS or FAIL" {
			return "FAIL", nil
		}
		return "bad answer", nil
	})

	if err := RunExperiment(context.Background(), db, id, mockChat); err != nil {
		t.Fatalf("RunExperiment: %v", err)
	}

	exp, _ := dal.GetExperiment(context.Background(), db, id)
	if exp.Status != "done" {
		t.Errorf("expected status 'done', got %q", exp.Status)
	}
	if exp.TotalScore == nil || *exp.TotalScore != 0 {
		t.Errorf("expected score 0, got %v", exp.TotalScore)
	}
}

func TestRunExperiment_ChatError(t *testing.T) {
	db := setupTestDB(t)
	id := seedExperiment(t, db, []string{"query one"})

	mockChat := ChatFunc(func(ctx context.Context, system, user string) (string, error) {
		return "", errors.New("ollama unavailable")
	})

	err := RunExperiment(context.Background(), db, id, mockChat)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	exp, _ := dal.GetExperiment(context.Background(), db, id)
	if exp.Status != "failed" {
		t.Errorf("expected status 'failed', got %q", exp.Status)
	}
}
