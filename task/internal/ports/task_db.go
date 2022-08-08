package ports

import (
	"context"

	"gitlab.com/g6834/team26/task/internal/domain/models"
)

type TaskDB interface {
	List(ctx context.Context, login string) ([]*models.Task, error)
	Run(ctx context.Context, t *models.Task) error
	Update(ctx context.Context, id, login, name, text string) error
	Delete(ctx context.Context, login, id string) error
	Approve(ctx context.Context, login, id, approvalLogin string) error
	Decline(ctx context.Context, login, id, approvalLogin string) error
	GetMessagesToSend(ctx context.Context) (map[int]models.KafkaAnalyticMessage, error)
	GetEmailsToSend(ctx context.Context) ([]models.Email, error)
	UpdateMessageStatus(ctx context.Context, id int) error
	ChangeEmailStatusAndSendMessage(ctx context.Context, e models.Email, result bool) error
}
