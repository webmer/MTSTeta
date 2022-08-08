package ports

import (
	"context"

	"gitlab.com/g6834/team26/task/internal/domain/models"
)

type Task interface {
	ListTasks(ctx context.Context, login string) ([]*models.Task, error)
	RunTask(ctx context.Context, createdTask *models.Task) error
	UpdateTask(ctx context.Context, id, login, name, text string) error
	DeleteTask(ctx context.Context, login, id string) error
	ApproveTask(ctx context.Context, login, id, approvalLogin string) error
	DeclineTask(ctx context.Context, login, id, approvalLogin string) error
	GrpcAuth
	StartMessageSender(ctx context.Context)
	StartEmailSender(ctx context.Context)
	Stop() error
}
