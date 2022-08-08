package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	e "gitlab.com/g6834/team26/task/internal/domain/errors"
	"gitlab.com/g6834/team26/task/internal/ports"
)

type Cookie struct {
	Name       string
	Value      string
	Expiration time.Time
}

func (s *Server) CheckProfiling() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if s.config.Server.Profiling {
				next.ServeHTTP(w, r)
			} else {
				http.Error(w, "{\"error\": \"pprof is off\"}", http.StatusServiceUnavailable)
			}
		}

		return http.HandlerFunc(fn)
	}
}

func (s *Server) setCookie(w http.ResponseWriter, c Cookie) {
	cookie := http.Cookie{
		Name:     c.Name,
		Value:    c.Value,
		Path:     "/",
		HttpOnly: true,
		Expires:  c.Expiration,
	}

	http.SetCookie(w, &cookie)
}

func (s *Server) ValidateTokens() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			tokens := ports.TokenPair{}
			accessToken, err := r.Cookie("access_token")
			if err != nil {
				tokens.AccessToken.Value = ""
			} else {
				tokens.AccessToken.Value = accessToken.Value
			}

			refreshToken, err := r.Cookie("refresh_token")
			if err != nil {
				http.Error(w, e.ErrAuthFailed.Error(), http.StatusForbidden)
				return
			}
			tokens.RefreshToken.Value = refreshToken.Value

			ctx := r.Context()

			grpcResponse, err := s.task.Validate(ctx, tokens)
			if errors.Is(err, e.ErrAuthFailed) || !grpcResponse.Result {
				http.Error(w, e.ErrAuthFailed.Error(), http.StatusForbidden)
				return
			} else if err != nil {
				http.Error(w, e.JsonErrWrapper{E: err.Error()}.Error(), http.StatusInternalServerError)
				return
			}

			if grpcResponse.RefreshToken != nil && grpcResponse.AccessToken != nil {
				s.setCookie(w, Cookie{
					Name:       "access_token",
					Value:      grpcResponse.AccessToken.GetValue(),
					Expiration: time.Unix(grpcResponse.AccessToken.GetExpires(), 0),
				})
				s.setCookie(w, Cookie{
					Name:       "refresh_token",
					Value:      grpcResponse.RefreshToken.GetValue(),
					Expiration: time.Unix(grpcResponse.RefreshToken.GetExpires(), 0),
				})
			}

			ctx = context.WithValue(ctx, "login", grpcResponse.Login)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
