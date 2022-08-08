package ports

import (
	"context"

	"gitlab.com/g6834/team26/task/internal/domain/models"
)

type EmailSender interface {
	StartEmailWorkers(ctx context.Context)
	SendEmail(e models.Email) error
	PushEmailToChan(e models.Email)
	GetEmailResultChan() chan map[models.Email]bool
}
