package postgres

import (
	"context"
	"gitlab.com/g6834/team26/analytic/internal/domain/models"
	"time"
)

func (pdb *PostgresDatabase) GetMessage(ctx context.Context, uuid string) (m *models.MessageDB, err error) {
	_, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	m = &models.MessageDB{}

	taskQuery := `SELECT "uuid", "task_uuid", "date_create", "type", "value" FROM "message" WHERE "uuid" = $1`
	taskRow := pdb.psqlClient.QueryRow(taskQuery, uuid)

	err = taskRow.Scan(&m.UUID, &m.TaskUUID, &m.DateCreate, &m.Type, &m.Value)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (pdb *PostgresDatabase) AddMessage(ctx context.Context, t *models.MessageDB) error {
	_, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := pdb.psqlClient.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	messageQuery := `INSERT INTO "message" ("task_uuid", "uuid", "date_create", "type", "value") VALUES ($1, $2, $3, $4, $5)`
	_, err = tx.ExecContext(ctx, messageQuery, t.TaskUUID, t.UUID, t.DateCreate, t.Type, t.Value)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
