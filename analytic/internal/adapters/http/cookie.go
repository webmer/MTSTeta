package http

import (
	"gitlab.com/g6834/team26/analytic/internal/domain/models"
	"net/http"
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
