package ports

import (
	"context"
	"gitlab.com/g6834/team26/analytic/internal/domain/models"
)

type Analytic interface {
	ApprovedTasks(ctx context.Context, login string) (int, error)
	DeclinedTasks(ctx context.Context, login string) (int, error)
	TotalTimeTasks(ctx context.Context, login string) ([]*models.Task, error)
	MessageIsExist(ctx context.Context, m *models.Message) (bool, error)
	ActionTask(ctx context.Context, m *models.Message) error
	GrpcAuth
}
