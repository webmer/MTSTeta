package postgres

import (
	"context"
	"database/sql"
	"errors"
	"gitlab.com/g6834/team26/analytic/internal/domain/models"
	"strings"
	"time"
)

func (pdb *PostgresDatabase) GetCountTasksByUserStatus(ctx context.Context, login string, status bool) (count int, err error) {
	_, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	taskQuery := `SELECT COUNT(*) FROM "task" WHERE "login" = $1 AND "status" = $2`
	taskCount := pdb.psqlClient.QueryRow(taskQuery, login, status)
	err = taskCount.Scan(&count)
	if err != nil {
		return
	}

	return
}

func (pdb *PostgresDatabase) GetTotalTimeTasks(ctx context.Context, login string) ([]*models.Task, error) {
	_, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var result []*models.Task

	taskQuery := `SELECT "uuid", "date_create", "date_action", "status" FROM "task" WHERE "login" = $1 AND "status" is not null`
	taskRows, err := pdb.psqlClient.Query(taskQuery, login)
	if err != nil {
		return nil, err
		// return nil, fmt.Errorf("no user with such login")
	}
	defer taskRows.Close()

	for taskRows.Next() {
		var UUID string
		var DateCreate time.Time
		var DateAction sql.NullTime
		var Status sql.NullString
		var task models.Task

		err := taskRows.Scan(&UUID, &DateCreate, &DateAction, &Status)
		if err != nil {
			return nil, err
		}
		task.UUID = strings.TrimSpace(UUID)
		if Status.Valid {
			task.Status = strings.TrimSpace(Status.String)
		} else {
			task.Status = ""
		}
		if DateAction.Valid {
			task.TotalTime = (int)(DateAction.Time.Sub(DateCreate).Seconds())
		} else {
			task.TotalTime = 0
		}
		result = append(result, &task)
	}
	return result, nil
}

func (pdb *PostgresDatabase) GetTask(ctx context.Context, uuid string) (task *models.TaskDB, err error) {
	_, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	task = &models.TaskDB{}

	taskQuery := `SELECT "uuid", "login", "date_create", "date_action", "status" FROM "task" WHERE "uuid" = $1`
	taskRow := pdb.psqlClient.QueryRow(taskQuery, uuid)

	err = taskRow.Scan(&task.UUID, &task.Login, &task.DateCreate, &task.DateAction, &task.Status)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (pdb *PostgresDatabase) AddTask(ctx context.Context, t *models.TaskDB) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := pdb.psqlClient.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	taskQuery := `INSERT INTO "task" ("uuid", "login", "date_create") VALUES ($1, $2, $3)`
	_, err = tx.ExecContext(ctx, taskQuery, t.UUID, t.Login, t.DateCreate)
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

func (pdb *PostgresDatabase) DateActionTask(ctx context.Context, t *models.TaskDB) error {
	_, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `UPDATE "task" SET "date_action" = $1 WHERE "uuid" = $2`
	result, err := pdb.psqlClient.Exec(query, t.DateAction, t.UUID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("not found")
	}
	return nil
}

func (pdb *PostgresDatabase) CompleteTask(ctx context.Context, t *models.TaskDB) error {
	_, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `UPDATE "task" SET "date_action" = $1, "status" = $2 WHERE "uuid" = $3`
	result, err := pdb.psqlClient.Exec(query, t.DateAction, t.Status, t.UUID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("not found")
	}
	return nil
}
