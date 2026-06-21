package dal

import (
	"context"
	"database/sql"
	"time"
)

type Dataset struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type DatasetCreated struct {
	Dataset
	QueryCount int `json:"query_count"`
}

func ListDatasets(ctx context.Context, db *sql.DB) ([]Dataset, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, name, created_at FROM datasets ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	datasets := []Dataset{}
	for rows.Next() {
		var d Dataset
		if err := rows.Scan(&d.ID, &d.Name, &d.CreatedAt); err != nil {
			return nil, err
		}
		datasets = append(datasets, d)
	}
	return datasets, rows.Err()
}

func ListQueries(ctx context.Context, db *sql.DB, datasetID int64) ([]string, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT query FROM user_queries WHERE dataset_id = $1 ORDER BY id`, datasetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var queries []string
	for rows.Next() {
		var q string
		if err := rows.Scan(&q); err != nil {
			return nil, err
		}
		queries = append(queries, q)
	}
	return queries, rows.Err()
}

func CreateDataset(ctx context.Context, db *sql.DB, name string, queries []string) (DatasetCreated, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return DatasetCreated{}, err
	}
	defer tx.Rollback()

	var d Dataset
	err = tx.QueryRowContext(ctx,
		`INSERT INTO datasets (name) VALUES ($1) RETURNING id, name, created_at`, name).
		Scan(&d.ID, &d.Name, &d.CreatedAt)
	if err != nil {
		return DatasetCreated{}, err
	}

	for _, q := range queries {
		if _, err = tx.ExecContext(ctx,
			`INSERT INTO user_queries (dataset_id, query) VALUES ($1, $2)`, d.ID, q); err != nil {
			return DatasetCreated{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return DatasetCreated{}, err
	}

	return DatasetCreated{Dataset: d, QueryCount: len(queries)}, nil
}
