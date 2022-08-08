package http

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"gitlab.com/g6834/team26/auth/internal/domain/models"
	"net/http"
)

const LoginFailed = "{\"Status\":\"error\"}"

func (s *Server) authHandlers() http.Handler {
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(s.ValidateCookieAuth())
		r.Post("/i", s.Validate)
	})

	r.Group(func(r chi.Router) {
		//r.Use(s.ValidateBasicAuth())
		r.Post("/login", s.Login)
	})
	r.Group(func(r chi.Router) {
		r.Use(s.ValidateCookieAuth())
		r.Post("/logout", s.Logout)
	})
	r.Group(func(r chi.Router) {
		r.Use(s.ValidateCookieAuth())
		r.Post("/create", s.Create)
	})
	return r
}

// Validate
// @ID Validate
// @tags auth
// @Summary Validate tokens
// @Description Validate tokens and refresh tokens if refresh token is valid
// @Security access_token
// @Security refresh_token
// @Body {object} models.Tokens true "Validate"
// @Success 200 {object} models.AuthResponse
// @Failure 401 {string} string "unauthorized"
// @Failure 403 {string} string "forbidden"
// @Failure 500 {string} string "internal error"
// @Router /i [post]
func (s *Server) Validate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := ctx.Value("user")

	req, ok := u.(CtxUser)
	if !ok {
		s.logger.Error().Msgf("wrong type %#v", req)
		http.Error(w, LoginFailed, http.StatusInternalServerError)
		return
	}

	res := models.AuthResponse{
		Status: "ok",
		Login:  req.Login,
	}

	if req.UpdTokens {
		tokens, err := s.auth.GenTokens(req.Login)
		if err != nil {
			//s.logger.Error().Msg(err.Error())
			http.Error(w, LoginFailed, http.StatusForbidden)
			return
		}

		res.AccessToken = tokens.AccessToken.Value
		res.RefreshToken = tokens.RefreshToken.Value
	}

	s.Redirect(w, r)

	err := s.AuthEncode(w, res)
	if err != nil {
		//s.logger.Error().Msg(err.Error())
		http.Error(w, LoginFailed, http.StatusForbidden)
		return
	}

	return
}

// Login
// @ID Login
// @tags auth
// @Summary Generate auth tokens.
// @Description Validate credentials, return access and refresh tokens.
// @Param data body models.AuthRequest true "login"
// @Success 200 {object} models.AuthResponse
// @Failure 403 {string} string "forbidden"
// @Failure 500 {string} string "internal error"
// @Router /login [post]
func (s *Server) Login(w http.ResponseWriter, r *http.Request) {
	ar, err := s.AuthDecode(r)

	if err != nil {
		http.Error(w, LoginFailed, http.StatusInternalServerError)
		s.logger.Error().Msg(err.Error())
		return
	}

	if ar.IsValid() {
		tokens, err := s.auth.Login(r.Context(), ar.Login, ar.Password)
		if err != nil {
			http.Error(w, LoginFailed, http.StatusForbidden)
			return
		}

		s.setCookie(w, models.Cookie{
			Name:       s.config.Server.AccessCookie,
			Value:      tokens.AccessToken.Value,
			Expiration: tokens.AccessToken.Expires,
		})
		s.setCookie(w, models.Cookie{
			Name:       s.config.Server.RefreshCookie,
			Value:      tokens.RefreshToken.Value,
			Expiration: tokens.RefreshToken.Expires,
		})

		s.Redirect(w, r)

		res := models.AuthResponse{
			Status:       "ok",
			AccessToken:  tokens.AccessToken.Value,
			RefreshToken: tokens.RefreshToken.Value,
		}
		err = s.AuthEncode(w, res)
		if err != nil {
			http.Error(w, LoginFailed, http.StatusForbidden)
			s.logger.Error().Msg(err.Error())
			return
		}

		return
	}
	http.Error(w, LoginFailed, http.StatusForbidden)
}

// Create
// @ID Create
// @tags create
// @Security access_token
// @Security refresh_token
// @Summary Create user for db.
// @Description Create user for db.
// @Success 200 {object} models.AuthResponse
// @Failure 401 {string} string "unauthorized"
// @Failure 403 {string} string "forbidden"
// @Failure 500 {string} string "internal error"
// @Router /create [post]
func (s *Server) Create(w http.ResponseWriter, r *http.Request) {
	ar, err := s.AuthDecode(r)

	if err != nil {
		http.Error(w, LoginFailed, http.StatusInternalServerError)
		s.logger.Error().Msg(err.Error())
		return
	}

	if ar.IsValid() {
		err := s.auth.Create(r.Context(), ar.Login, ar.Password)
		if err != nil {
			http.Error(w, LoginFailed, http.StatusForbidden)
			s.logger.Error().Msg(err.Error())
			return
		}

		return
	}
	http.Error(w, LoginFailed, http.StatusForbidden)
}

// Logout
// @ID Logout
// @tags logout
// @Security access_token
// @Security refresh_token
// @Summary Logout user.
// @Description Logout user, delete cookie tokens.
// @Success 200 {object} models.AuthResponse
// @Failure 401 {string} string "unauthorized"
// @Failure 403 {string} string "forbidden"
// @Failure 500 {string} string "internal error"
// @Router /logout [post]
func (s *Server) Logout(w http.ResponseWriter, r *http.Request) {
	accessC := s.deleteCookie(s.config.Server.AccessCookie)
	refreshC := s.deleteCookie(s.config.Server.RefreshCookie)

	s.setCookie(w, accessC)
	s.setCookie(w, refreshC)

	s.Redirect(w, r)

	res := models.AuthResponse{Status: "ok"}
	err := s.AuthEncode(w, res)
	if err != nil {
		http.Error(w, LoginFailed, http.StatusForbidden)
		s.logger.Error().Msg(err.Error())
		return
	}
}

func (s *Server) Redirect(w http.ResponseWriter, r *http.Request) {
	redirectURL := r.FormValue("redirect_url")
	if len(redirectURL) > 0 {
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
	}
}

func (s *Server) AuthDecode(r *http.Request) (ar *models.AuthRequest, err error) {
	ar = &models.AuthRequest{}

	e := json.NewDecoder(r.Body)
	err = e.Decode(ar)
	if err != nil {
		return
	}

	return
}

func (s *Server) AuthEncode(w http.ResponseWriter, res models.AuthResponse) (err error) {
	resJson, err := json.Marshal(res)
	if err != nil {
		return
	}

	_, err = w.Write(resJson)
	if err != nil {
		return
	}

	return
}
