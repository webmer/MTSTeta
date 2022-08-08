package ports

import (
	"gitlab.com/g6834/team26/analytic/internal/domain/models"
	"gitlab.com/g6834/team26/analytic/pkg/api"
)

type GrpcAuth interface {
	Validate(tokens models.TokenPair) (*api.AuthResponse, error)
}
