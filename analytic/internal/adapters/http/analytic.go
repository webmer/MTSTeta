package http

import (
	"encoding/json"
	"gitlab.com/g6834/team26/analytic/internal/domain/models"
	"net/http"

	"github.com/go-chi/chi/v5"
	e "gitlab.com/g6834/team26/analytic/internal/domain/errors"
)

func (s *Server) analyticHandlers() http.Handler {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(s.ValidateTokens())
		r.Post("/approved_tasks", s.ApprovedTasks)
	})

	r.Group(func(r chi.Router) {
		r.Use(s.ValidateTokens())
		r.Post("/declined_tasks", s.DeclinedTasks)
	})

	r.Group(func(r chi.Router) {
		r.Use(s.ValidateTokens())
		r.Post("/total_time_tasks", s.TotalTimeTasks)
	})

	return r
}

// ApprovedTasks
// @ID ApprovedTasks
// @tags approved_tasks
// @Security access_token
// @Security refresh_token
// @Summary Get count approved task.
// @Description Get count approved task.
// @Success 200 {object} models.TaskCountResponse
// @Failure 400 {string} string "bad request
// @Failure 403 {string} string "forbidden"
// @Failure 500 {string} string "internal error"
// @Router /approved_tasks [post]
func (s *Server) ApprovedTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()

	login := ctx.Value("login").(string)

	t, err := s.analytic.ApprovedTasks(ctx, login)
	if err != nil {
		s.logger.Error().Msg(err.Error())
		http.Error(w, e.JsonErrWrapper{E: err.Error()}.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(models.TaskCountResponse{Count: t})
}

// DeclinedTasks
// @ID DeclinedTasks
// @tags declined_tasks
// @Security access_token
// @Security refresh_token
// @Summary Get count declined task.
// @Description Get count declined task.
// @Success 200 {object} models.TaskCountResponse
// @Failure 400 {string} string "bad request
// @Failure 403 {string} string "forbidden"
// @Failure 500 {string} string "internal error"
// @Router /declined_tasks [post]
func (s *Server) DeclinedTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()

	login := ctx.Value("login").(string)

	t, err := s.analytic.DeclinedTasks(ctx, login)
	if err != nil {
		s.logger.Error().Msg(err.Error())
		http.Error(w, e.JsonErrWrapper{E: err.Error()}.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(models.TaskCountResponse{Count: t})
}

// TotalTimeTasks
// @ID TotalTimeTasks
// @tags total_time_tasks
// @Security access_token
// @Security refresh_token
// @Summary Total waiting time of reactions for each task.
// @Description Total waiting time of reactions for each task.
// @Success 200 {object} models.TotalTimeTasksResponse
// @Failure 400 {string} string "bad request
// @Failure 403 {string} string "forbidden"
// @Failure 500 {string} string "internal error"
// @Router /total_time_tasks [post]
func (s *Server) TotalTimeTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()

	login := ctx.Value("login").(string)

	t, err := s.analytic.TotalTimeTasks(ctx, login)
	if err != nil {
		s.logger.Error().Msg(err.Error())
		http.Error(w, e.JsonErrWrapper{E: err.Error()}.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(models.TotalTimeTasksResponse{Tasks: t})
}
