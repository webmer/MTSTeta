package ports

import (
	"context"
	"time"

	"gitlab.com/g6834/team26/task/pkg/api"
)

type TokenPair struct {
	AccessToken  TokenPairVal
	RefreshToken TokenPairVal
}

type TokenPairVal struct {
	Value   string
	Expires time.Time
}

type GrpcAuth interface {
	Validate(ctx context.Context, tokens TokenPair) (*api.AuthResponse, error)
}
