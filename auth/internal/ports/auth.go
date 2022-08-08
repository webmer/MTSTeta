package ports

import (
	"context"

	"gitlab.com/g6834/team26/auth/internal/domain/models"
)

type Auth interface {
	//Info(ctx context.Context, login string) (*models.User, error)
	Validate(ctx context.Context, tokens models.TokenPair) (string, bool, error)
	Login(ctx context.Context, login, password string) (models.TokenPair, error)
	Create(ctx context.Context, login, password string) error
	GenTokens(login string) (models.TokenPair, error)
}
