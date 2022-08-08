package ports

import (
	"context"

	"gitlab.com/g6834/team26/task/internal/domain/models"
)

type TaskAnalyticSender interface {
	ActionTask(ctx context.Context, m models.KafkaAnalyticMessage) error
}
