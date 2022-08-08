package http

import (
	"context"
	"gitlab.com/g6834/team26/analytic/internal/domain/models"
	"net/http"
	"time"
)

func (s *Server) CheckProfiling() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if s.config.Server.Profiling {
				next.ServeHTTP(w, r)
			} else {
				http.Error(w, "", http.StatusNotFound)
			}
		}

		return http.HandlerFunc(fn)
	}
}

func (s *Server) ValidateTokens() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokens := models.TokenPair{}
			accessToken, err := r.Cookie(s.config.Server.AccessCookie)
			if err != nil {
				tokens.AccessToken.Value = ""
			} else {
				tokens.AccessToken.Value = accessToken.Value
			}

			refreshToken, err := r.Cookie(s.config.Server.RefreshCookie)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}
			tokens.RefreshToken.Value = refreshToken.Value
			grpcResponse, err := s.analytic.Validate(tokens)
			if err != nil || !grpcResponse.Result {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}

			if grpcResponse.AccessToken != nil && grpcResponse.RefreshToken != nil {
				s.setCookie(w, models.Cookie{
					Name:       s.config.Server.AccessCookie,
					Value:      grpcResponse.AccessToken.Value,
					Expiration: time.Unix(grpcResponse.AccessToken.Expires, 0),
				})
				s.setCookie(w, models.Cookie{
					Name:       s.config.Server.RefreshCookie,
					Value:      grpcResponse.RefreshToken.Value,
					Expiration: time.Unix(grpcResponse.RefreshToken.Expires, 0),
				})
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, "login", grpcResponse.Login)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
