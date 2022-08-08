package postgres

import (
	"context"
	"database/sql"
	_ "github.com/lib/pq"
	"time"
)

type PostgresDatabase struct {
	psqlClient *sql.DB
}

func New(ctx context.Context, pgconn string) (*PostgresDatabase, error) {
	_, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	db, err := sql.Open("postgres", pgconn+"?sslmode=disable&search_path=analytic")

	if err != nil {
		return nil, err
	}
	return &PostgresDatabase{psqlClient: db}, nil
}

func (pdb *PostgresDatabase) Stop(ctx context.Context) error {
	err := pdb.psqlClient.Close()
	if err != nil {
		return err
	}
	return nil
}
