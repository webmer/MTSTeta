package grpc

import (
	"gitlab.com/g6834/team26/auth/internal/domain/models"
	"gitlab.com/g6834/team26/auth/internal/ports"
	"gitlab.com/g6834/team26/auth/pkg/api"
	"golang.org/x/net/context"
)

type AuthServer struct {
	authS ports.Auth
	api.UnimplementedAuthServer
}

func (a *AuthServer) VerifyToken(ctx context.Context, auth *api.AuthRequest) (r *api.AuthResponse, err error) {
	r = &api.AuthResponse{
		Result: false,
	}

	login, upd, err := a.authS.Validate(context.Background(), models.TokenPair{
		AccessToken:  models.TokenPairVal{Value: auth.GetAccessToken()},
		RefreshToken: models.TokenPairVal{Value: auth.GetRefreshToken()},
	})
	if err != nil {
		return r, nil
	}

	if upd {
		tokens, err := a.authS.GenTokens(login)
		if err != nil {
			return r, nil
		}

		r.AccessToken = &api.Token{
			Value:   tokens.AccessToken.Value,
			Expires: tokens.AccessToken.Expires.Unix(),
		}
		r.RefreshToken = &api.Token{
			Value:   tokens.RefreshToken.Value,
			Expires: tokens.RefreshToken.Expires.Unix(),
		}
	}

	r.Login = login
	r.Result = true

	return r, nil
}
