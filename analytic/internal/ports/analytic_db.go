package ports

import (
	"context"
	"gitlab.com/g6834/team26/analytic/internal/domain/models"
)

type AnalyticDB interface {
	GetCountTasksByUserStatus(ctx context.Context, login string, status bool) (count int, err error)
	GetTotalTimeTasks(ctx context.Context, login string) (task []*models.Task, err error)

	GetTask(ctx context.Context, uuid string) (task *models.TaskDB, err error)
	GetMessage(ctx context.Context, uuid string) (message *models.MessageDB, err error)
	AddTask(ctx context.Context, t *models.TaskDB) error
	AddMessage(ctx context.Context, t *models.MessageDB) error
	DateActionTask(ctx context.Context, t *models.TaskDB) error
	CompleteTask(ctx context.Context, t *models.TaskDB) error
}
