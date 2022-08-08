package ports

import (
	"context"

	"gitlab.com/g6834/team26/auth/internal/domain/models"
)

type UserStorage interface {
	Create(ctx context.Context, login, password string) error
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
	//FindOne(ctx context.Context, id string) (*models.User, error)
}
