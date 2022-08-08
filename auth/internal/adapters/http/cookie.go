package http

import (
	"gitlab.com/g6834/team26/auth/internal/domain/models"
	"net/http"
	"time"
)

func (s *Server) setCookie(w http.ResponseWriter, c models.Cookie) {
	cookie := http.Cookie{
		Name:     c.Name,
		Value:    c.Value,
		Path:     "/",
		HttpOnly: true,
		Expires:  c.Expiration,
	}

	http.SetCookie(w, &cookie)
}

func (s *Server) deleteCookie(name string) models.Cookie {
	return models.Cookie{
		Name:       name,
		Value:      "",
		Expiration: time.Unix(0, 0),
	}
}
