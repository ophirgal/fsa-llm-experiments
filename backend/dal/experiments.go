package dal

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var ErrAlreadyStarted = errors.New("experiment already started")

type Experiment struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	DatasetID  int64      `json:"dataset_id"`
	Status     string     `json:"status"`
	TotalScore *float64   `json:"total_score"`
	StartTime  *time.Time `json:"start_time"`
	EndTime    *time.Time `json:"end_time"`
	CreatedAt  time.Time  `json:"created_at"`
}

type ExperimentDetail struct {
	Experiment
	Prompts     []string `json:"prompts"`
	JudgePrompt string   `json:"judge_prompt"`
}

func ListExperiments(ctx context.Context, db *sql.DB) ([]Experiment, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, name, dataset_id, status, total_score, start_time, end_time, created_at
		 FROM experiments ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	experiments := []Experiment{}
	for rows.Next() {
		var e Experiment
		if err := rows.Scan(&e.ID, &e.Name, &e.DatasetID, &e.Status,
			&e.TotalScore, &e.StartTime, &e.EndTime, &e.CreatedAt); err != nil {
			return nil, err
		}
		experiments = append(experiments, e)
	}
	return experiments, rows.Err()
}

func GetExperiment(ctx context.Context, db *sql.DB, id int64) (ExperimentDetail, error) {
	var e Experiment
	err := db.QueryRowContext(ctx,
		`SELECT id, name, dataset_id, status, total_score, start_time, end_time, created_at
		 FROM experiments WHERE id = $1`, id).
		Scan(&e.ID, &e.Name, &e.DatasetID, &e.Status,
			&e.TotalScore, &e.StartTime, &e.EndTime, &e.CreatedAt)
	if err != nil {
		return ExperimentDetail{}, err
	}

	prompts, err := listPrompts(ctx, db, id)
	if err != nil {
		return ExperimentDetail{}, err
	}

	judgePrompt, err := getJudgePrompt(ctx, db, id)
	if err != nil {
		return ExperimentDetail{}, err
	}

	return ExperimentDetail{Experiment: e, Prompts: prompts, JudgePrompt: judgePrompt}, nil
}

func CreateExperiment(ctx context.Context, db *sql.DB, name string, datasetID int64, prompts []string, judgePrompt string) (ExperimentDetail, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return ExperimentDetail{}, err
	}
	defer tx.Rollback()

	var e Experiment
	err = tx.QueryRowContext(ctx,
		`INSERT INTO experiments (name, dataset_id) VALUES ($1, $2)
		 RETURNING id, name, dataset_id, status, total_score, start_time, end_time, created_at`,
		name, datasetID).
		Scan(&e.ID, &e.Name, &e.DatasetID, &e.Status,
			&e.TotalScore, &e.StartTime, &e.EndTime, &e.CreatedAt)
	if err != nil {
		return ExperimentDetail{}, err
	}

	for _, p := range prompts {
		if _, err = tx.ExecContext(ctx,
			`INSERT INTO experiment_prompts (experiment_id, prompt) VALUES ($1, $2)`, e.ID, p); err != nil {
			return ExperimentDetail{}, err
		}
	}

	if _, err = tx.ExecContext(ctx,
		`INSERT INTO experiment_judge_prompt (experiment_id, prompt) VALUES ($1, $2)`, e.ID, judgePrompt); err != nil {
		return ExperimentDetail{}, err
	}

	if err := tx.Commit(); err != nil {
		return ExperimentDetail{}, err
	}

	return ExperimentDetail{Experiment: e, Prompts: prompts, JudgePrompt: judgePrompt}, nil
}

func StartExperiment(ctx context.Context, db *sql.DB, id int64) (Experiment, error) {
	var e Experiment
	err := db.QueryRowContext(ctx,
		`UPDATE experiments
		 SET status = 'in progress', start_time = NOW()
		 WHERE id = $1 AND status = 'ready'
		 RETURNING id, name, dataset_id, status, total_score, start_time, end_time, created_at`, id).
		Scan(&e.ID, &e.Name, &e.DatasetID, &e.Status,
			&e.TotalScore, &e.StartTime, &e.EndTime, &e.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Experiment{}, ErrAlreadyStarted
	}
	return e, err
}

func FinishExperiment(ctx context.Context, db *sql.DB, id int64, score float64, failed bool) error {
	status := "done"
	if failed {
		status = "failed"
	}
	_, err := db.ExecContext(ctx,
		`UPDATE experiments SET status = $1, end_time = NOW(), total_score = $2 WHERE id = $3`,
		status, score, id)
	return err
}

func listPrompts(ctx context.Context, db *sql.DB, experimentID int64) ([]string, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT prompt FROM experiment_prompts WHERE experiment_id = $1 ORDER BY id`, experimentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prompts []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		prompts = append(prompts, p)
	}
	return prompts, rows.Err()
}

func getJudgePrompt(ctx context.Context, db *sql.DB, experimentID int64) (string, error) {
	var prompt string
	err := db.QueryRowContext(ctx,
		`SELECT prompt FROM experiment_judge_prompt WHERE experiment_id = $1`, experimentID).
		Scan(&prompt)
	return prompt, err
}
