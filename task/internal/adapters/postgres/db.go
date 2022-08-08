package postgres

import (
	"context"
	"database/sql"
	"log"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
	e "gitlab.com/g6834/team26/task/internal/domain/errors"
	"gitlab.com/g6834/team26/task/internal/domain/models"
)

type PostgresDatabase struct {
	psqlClient *sql.DB
}

func New(ctx context.Context, pgconn string) (*PostgresDatabase, error) {
	// _, cancel := context.WithTimeout(ctx, 5*time.Second)
	// defer cancel()
	db, err := sql.Open("postgres", pgconn+"?sslmode=disable")

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

func (pdb *PostgresDatabase) List(ctx context.Context, login string) ([]*models.Task, error) {
	// _, cancel := context.WithTimeout(ctx, 5*time.Second)
	// defer cancel()
	var result []*models.Task

	taskQuery := `SELECT "uuid", "name", "text", "login", "status" FROM "tasks" WHERE "login" = $1`
	taskRows, err := pdb.psqlClient.Query(taskQuery, login)
	if err != nil {
		return nil, err
		// return nil, fmt.Errorf("no user with such login")
	}
	defer taskRows.Close()

	for taskRows.Next() {
		var task models.Task
		err := taskRows.Scan(&task.UUID, &task.Name, &task.Text, &task.InitiatorLogin, &task.Status)
		if err != nil {
			return nil, err
		}
		task.Name = strings.TrimSpace(task.Name)
		task.Text = strings.TrimSpace(task.Text)
		task.InitiatorLogin = strings.TrimSpace(task.InitiatorLogin)
		task.Status = strings.TrimSpace(task.Status)

		approvalQuery := `SELECT "approval_login", "approved", "sent", "n" FROM "approvals" WHERE "task_uuid" = $1`
		approvalRows, err := pdb.psqlClient.Query(approvalQuery, task.UUID)
		if err != nil {
			return nil, err
		}
		defer approvalRows.Close()

		for approvalRows.Next() {
			var approval models.Approval
			err := approvalRows.Scan(&approval.ApprovalLogin, &approval.Approved, &approval.Sent, &approval.N)
			if err != nil {
				return nil, err
			}
			approval.ApprovalLogin = strings.TrimSpace(approval.ApprovalLogin)
			task.Approvals = append(task.Approvals, &approval)
		}
		result = append(result, &task)
	}

	return result, nil
}

func (pdb *PostgresDatabase) Run(ctx context.Context, t *models.Task) error {
	// _, cancel := context.WithTimeout(ctx, 5*time.Second)
	// defer cancel()
	tx, err := pdb.psqlClient.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	taskQuery := `INSERT INTO "tasks" ("uuid", "name", "text", "login", "status") VALUES ($1, $2, $3, $4, $5)`
	_, err = tx.Exec(taskQuery, t.UUID, t.Name, t.Text, t.InitiatorLogin, t.Status)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, approval := range t.Approvals {
		approvalsQuery := `INSERT INTO "approvals" ("task_uuid", "approval_login", "n") VALUES ($1, $2, $3)`
		_, err = tx.Exec(approvalsQuery, t.UUID, approval.ApprovalLogin, approval.N)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = pdb.SaveMessage(tx, t.UUID, "run", t.InitiatorLogin, time.Now().Unix())
	if err != nil {
		tx.Rollback()
		return err
	}

	err = pdb.saveEmailMessage(tx, t.UUID, t.Approvals[0].ApprovalLogin, "approve")
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

func (pdb *PostgresDatabase) Update(ctx context.Context, id, login, name, text string) error {
	// _, cancel := context.WithTimeout(ctx, 5*time.Second)
	// defer cancel()
	var (
		query  string
		err    error
		result sql.Result
	)

	switch {
	case text == "" && name == "":
		log.Println("text == \"\" && name == \"\"")
		return e.ErrNothingToChange
	case text == "":
		log.Println("text == \"\"")
		query = `UPDATE "tasks" SET "name" = $1 WHERE "uuid" = $2 AND "login" = $3`
		result, err = pdb.psqlClient.Exec(query, name, id, login)
	case name == "":
		log.Println("name == \"\"")
		query = `UPDATE "tasks" SET "text" = $1 WHERE "uuid" = $2 AND "login" = $3`
		result, err = pdb.psqlClient.Exec(query, text, id, login)
	default:
		log.Println("default")
		query = `UPDATE "tasks" SET "name" = $1, "text" = $2 WHERE "uuid" = $3 AND "login" = $4`
		result, err = pdb.psqlClient.Exec(query, name, text, id, login)
	}
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return e.ErrNotFound
	}
	return nil
}

func (pdb *PostgresDatabase) Delete(ctx context.Context, login, id string) error {
	// _, cancel := context.WithTimeout(ctx, 5*time.Second)
	// defer cancel()
	tx, err := pdb.psqlClient.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	query := `UPDATE "tasks" SET "status" = $1 WHERE "uuid" = $2 AND "login" = $3`
	result, err := tx.Exec(query, "deleted", id, login)
	if err != nil {
		tx.Rollback()
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return e.ErrNotFound
	}

	err = pdb.SaveMessage(tx, id, "delete", "true", time.Now().Unix())
	if err != nil {
		tx.Rollback()
		return err
	}

	// отправляем всем письма в связи с удалением задачи
	err = pdb.sendInfoEmails(ctx, tx, id, pdb.getTaskInitiator(id), "deleted")
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

func (pdb *PostgresDatabase) Approve(ctx context.Context, login, id, approvalLogin string) error {
	// _, cancel := context.WithTimeout(ctx, 5*time.Second)
	// defer cancel()
	tx, err := pdb.psqlClient.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	err = pdb.checkTaskStatus(id)
	if err != nil {
		return err
	}

	err = pdb.checkApproval(id, approvalLogin)
	if err != nil {
		return err
	}

	query := `UPDATE "approvals" SET "approved" = $1 WHERE "task_uuid" = $2 AND "approval_login" = $3`
	result, err := tx.Exec(query, true, id, approvalLogin)
	if err != nil {
		tx.Rollback()
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return e.ErrNotFound
	}

	currPosition := pdb.getCurrentApprovalPosition(id, approvalLogin)
	maxPosition := pdb.getMaxApprovalPosition(id)

	// log.Println("currPosition -", currPosition)
	// log.Println("maxPosition -", maxPosition)

	err = pdb.SaveMessage(tx, id, "approve", "true", time.Now().Unix())
	if err != nil {
		tx.Rollback()
		return err
	}

	if currPosition == maxPosition {
		err = pdb.changeTaskStatus(tx, id, "completed")
		if err != nil {
			tx.Rollback()
			return err
		}
		err = pdb.SaveMessage(tx, id, "complete", "true", time.Now().Unix())
		if err != nil {
			tx.Rollback()
			return err
		}
		// если последний отправляем всем письма о завершении задачи
		err = pdb.sendInfoEmails(ctx, tx, id, pdb.getTaskInitiator(id), "completed")
		if err != nil {
			tx.Rollback()
			return err
		}
	} else {
		// если не последний, отправляем письмо следующему
		err = pdb.saveEmailMessage(tx, id, pdb.getNextApproval(id, currPosition+1), "approve")
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (pdb *PostgresDatabase) Decline(ctx context.Context, login, id, approvalLogin string) error {
	// _, cancel := context.WithTimeout(ctx, 5*time.Second)
	// defer cancel()
	tx, err := pdb.psqlClient.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	err = pdb.checkTaskStatus(id)
	if err != nil {
		return err
	}

	err = pdb.checkApproval(id, approvalLogin)
	if err != nil {
		return err
	}

	query := `UPDATE "approvals" SET "approved" = $1 WHERE "task_uuid" = $2 AND "approval_login" = $3`
	result, err := tx.Exec(query, false, id, approvalLogin)
	if err != nil {
		tx.Rollback()
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return e.ErrNotFound
	}

	err = pdb.changeTaskStatus(tx, id, "declined")
	if err != nil {
		tx.Rollback()
		return err
	}

	err = pdb.SaveMessage(tx, id, "approve", "false", time.Now().Unix())
	if err != nil {
		tx.Rollback()
		return err
	}

	// отправляем письма всем участникам в связи с отклонением задачи одним из участников
	err = pdb.sendInfoEmails(ctx, tx, id, pdb.getTaskInitiator(id), "declined")
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

func (pdb *PostgresDatabase) checkApproval(id, approvalLogin string) error {
	var approval models.Approval
	approvalQuery := `SELECT "approved" FROM "approvals" WHERE "task_uuid" = $1 AND "approval_login" = $2`
	approvalRow := pdb.psqlClient.QueryRow(approvalQuery, id, approvalLogin)
	approvalRow.Scan(&approval.Approved)
	if approval.Approved.Valid {
		return e.ErrApprovalHasBeenDone
	}
	return nil
}

func (pdb *PostgresDatabase) checkTaskStatus(id string) error {
	var task models.Task
	taskQuery := `SELECT "status" FROM "tasks" WHERE "uuid" = $1`
	taskRow := pdb.psqlClient.QueryRow(taskQuery, id)
	taskRow.Scan(&task.Status)
	task.Status = strings.TrimSpace(task.Status)
	if task.Status != "created" {
		return e.ErrTaskNotAvailableForApproval
	}
	return nil
}

func (pdb *PostgresDatabase) changeTaskStatus(tx *sql.Tx, id, status string) error {
	query := `UPDATE "tasks" SET "status" = $1 WHERE "uuid" = $2`
	_, err := tx.Exec(query, status, id)
	if err != nil {
		return err
	}
	return nil
}

func (pdb *PostgresDatabase) getCurrentApprovalPosition(id, approvalLogin string) int {
	var currentApprovalPosition int
	maxApprovalQuery := `SELECT "n" FROM "approvals" WHERE "task_uuid" = $1 AND "approval_login" = $2 ORDER BY "n" DESC LIMIT 1`
	maxApprovalRow := pdb.psqlClient.QueryRow(maxApprovalQuery, id, approvalLogin)
	maxApprovalRow.Scan(&currentApprovalPosition)
	return currentApprovalPosition
}

func (pdb *PostgresDatabase) getMaxApprovalPosition(id string) int {
	var maxApprovalPosition int
	maxApprovalQuery := `SELECT "n" FROM "approvals" WHERE "task_uuid" = $1 ORDER BY "n" DESC LIMIT 1`
	maxApprovalRow := pdb.psqlClient.QueryRow(maxApprovalQuery, id)
	maxApprovalRow.Scan(&maxApprovalPosition)
	return maxApprovalPosition
}

func (pdb *PostgresDatabase) SaveMessage(tx *sql.Tx, id, t, v string, aT int64) error {
	query := `INSERT INTO "outbox" ("task_uuid", "action_timestamp", "type", "value") VALUES ($1, $2, $3, $4)`
	_, err := tx.Exec(query, id, aT, t, v)
	if err != nil {
		return err
	}
	return nil
}

func (pdb *PostgresDatabase) GetMessagesToSend(ctx context.Context) (map[int]models.KafkaAnalyticMessage, error) {
	messages := make(map[int]models.KafkaAnalyticMessage)

	getMessagesQuery := `SELECT "id", "task_uuid", "action_timestamp", "type", "value" FROM "outbox" WHERE "sent" IS NULL`
	messagesRows, err := pdb.psqlClient.Query(getMessagesQuery)
	if err != nil {
		return nil, err
	}
	defer messagesRows.Close()

	for messagesRows.Next() {
		var message models.KafkaAnalyticMessage
		var id int
		err := messagesRows.Scan(&id, &message.UUID, &message.Timestamp, &message.Type, &message.Value)
		if err != nil {
			return nil, err
		}
		message.Type = strings.TrimSpace(message.Type)
		message.Value = strings.TrimSpace(message.Value)
		messages[id] = message
	}

	return messages, nil
}

func (pdb *PostgresDatabase) UpdateMessageStatus(ctx context.Context, id int) error {
	query := `UPDATE "outbox" SET "sent" = $1 WHERE "id" = $2`
	_, err := pdb.psqlClient.Exec(query, true, id)
	if err != nil {
		return err
	}
	return nil
}

func (pdb *PostgresDatabase) saveEmailMessage(tx *sql.Tx, id, reciever, t string) error {
	query := `INSERT INTO "outbox_email" ("task_uuid", "reciever", "type") VALUES ($1, $2, $3)`
	_, err := tx.Exec(query, id, reciever, t)
	if err != nil {
		return err
	}
	return nil
}

func (pdb *PostgresDatabase) GetEmailsToSend(ctx context.Context) ([]models.Email, error) {
	emails := make([]models.Email, 0)

	getEmailsQuery := `SELECT "id", "task_uuid", "reciever", "type" FROM "outbox_email" WHERE "sent" IS NULL`
	emailsRows, err := pdb.psqlClient.Query(getEmailsQuery)
	if err != nil {
		return nil, err
	}
	defer emailsRows.Close()

	for emailsRows.Next() {
		var email models.Email
		err := emailsRows.Scan(&email.Id, &email.TaskUUID, &email.Reciever, &email.Type)
		if err != nil {
			return nil, err
		}
		email.Reciever = strings.TrimSpace(email.Reciever)
		email.Type = strings.TrimSpace(email.Type)
		emails = append(emails, email)
	}

	return emails, nil
}

func (pdb *PostgresDatabase) UpdateEmailSendStatus(ctx context.Context, id int) error {
	query := `UPDATE "outbox_email" SET "sent" = $1 WHERE "id" = $2`
	_, err := pdb.psqlClient.Exec(query, true, id)
	if err != nil {
		return err
	}
	return nil
}

func (pdb *PostgresDatabase) ChangeEmailStatusAndSendMessage(ctx context.Context, e models.Email, result bool) error {
	tx, err := pdb.psqlClient.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	err = pdb.UpdateEmailSendStatus(ctx, e.Id)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = pdb.SaveMessage(tx, e.TaskUUID, "send", strconv.FormatBool(result), time.Now().Unix())
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

func (pdb *PostgresDatabase) getApprovals(ctx context.Context, id string) ([]string, error) {
	approvals := make([]string, 0)

	getApprovalsQuery := `SELECT "approval_login" FROM "approvals" WHERE "task_uuid" = $1`
	approvalsRows, err := pdb.psqlClient.Query(getApprovalsQuery, id)
	if err != nil {
		return nil, err
	}
	defer approvalsRows.Close()

	for approvalsRows.Next() {
		var approval string
		err := approvalsRows.Scan(&approval)
		if err != nil {
			return nil, err
		}
		approval = strings.TrimSpace(approval)
		approvals = append(approvals, approval)
	}

	return approvals, nil
}

func (pdb *PostgresDatabase) sendInfoEmails(ctx context.Context, tx *sql.Tx, id, login, t string) error {
	err := pdb.saveEmailMessage(tx, id, login, t)
	if err != nil {
		tx.Rollback()
		return err
	}

	approvals, err := pdb.getApprovals(ctx, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, approval := range approvals {
		err = pdb.saveEmailMessage(tx, id, approval, t)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return nil
}

func (pdb *PostgresDatabase) getNextApproval(id string, n int) string {
	var nextApproval string
	nextApprovalQuery := `SELECT "approval_login" FROM "approvals" WHERE "task_uuid" = $1 AND "n" = $2`
	nextApprovalRow := pdb.psqlClient.QueryRow(nextApprovalQuery, id, n)
	nextApprovalRow.Scan(&nextApproval)
	return nextApproval
}

func (pdb *PostgresDatabase) getTaskInitiator(id string) string {
	var taskInitiator string
	taskInitiatorQuery := `SELECT "login" FROM "tasks" WHERE "uuid" = $1`
	taskInitiatorRow := pdb.psqlClient.QueryRow(taskInitiatorQuery, id)
	taskInitiatorRow.Scan(&taskInitiator)
	return taskInitiator
}
