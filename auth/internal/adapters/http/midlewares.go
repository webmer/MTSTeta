package http

import (
	"context"
	"encoding/json"
	"net/http"

	"gitlab.com/g6834/team26/auth/internal/domain/models"
)

const (
	errInvalidToken = "invalid token"
)

func (s *Server) ValidateCookieAuth() func(next http.Handler) http.Handler {
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
				http.Error(w, errInvalidToken, http.StatusUnauthorized)
				return
			}
			tokens.RefreshToken.Value = refreshToken.Value

			login, upd, err := s.auth.Validate(r.Context(), tokens)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			ctx := r.Context()

			ctx = context.WithValue(ctx, "user", CtxUser{Login: login, UpdTokens: upd})

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (s *Server) ValidateBasicAuth() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ar, err := s.AuthDecode(r)

			if err != nil {
				s.logger.Error().Msg(err.Error())
			}

			if ar != nil && ar.IsValid() {
				ar.Auth = true
			}
			r = r.WithContext(context.WithValue(r.Context(), "user", ar))

			ctx := r.Context()

			//ctx = context.WithValue(ctx, "user", 1)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (s *Server) ValidateTokenAuth() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ar := &models.AuthTokenRequest{}

			e := json.NewDecoder(r.Body)
			err := e.Decode(ar)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			login, upd, err := s.auth.Validate(r.Context(), models.TokenPair{
				AccessToken:  models.TokenPairVal{Value: ar.AccessToken},
				RefreshToken: models.TokenPairVal{Value: ar.RefreshToken},
			})
			if err != nil {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			ctx := r.Context()

			ctx = context.WithValue(ctx, "user", CtxUser{Login: login, UpdTokens: upd})

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

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
