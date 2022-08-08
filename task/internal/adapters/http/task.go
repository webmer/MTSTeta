package http

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	e "gitlab.com/g6834/team26/task/internal/domain/errors"
	"gitlab.com/g6834/team26/task/internal/domain/models"
)

// @title Сервис создания и согласования задач
// @version 1.0
// @description Сервис для создания и согласования задач и последующей отправкой писем последовательно всем участвующим лицам.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:3000
// @BasePath /task/v1

// @securityDefinitions.apikey access_token
// @in header
// @name access_token
// @securityDefinitions.apikey refresh_token
// @in header
// @name refresh_token

func (s *Server) taskHandlers() http.Handler {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(s.ValidateTokens())
		r.Post("/tasks/run", s.RunTaskHandler)
	})

	r.Group(func(r chi.Router) {
		r.Use(s.ValidateTokens())
		r.Get("/tasks/", s.GetTasksListHandler)
	})

	r.Group(func(r chi.Router) {
		r.Use(s.ValidateTokens())
		r.Put("/tasks/{taskID}", s.UpdateTaskHandler)
	})

	r.Group(func(r chi.Router) {
		r.Use(s.ValidateTokens())
		r.Delete("/tasks/{taskID}", s.DeleteTaskHandler)
	})

	r.Group(func(r chi.Router) {
		r.Use(s.ValidateTokens())
		r.Post("/tasks/{taskID}/approve/{approvalLogin}", s.ApproveTaskHandler)
	})

	r.Group(func(r chi.Router) {
		r.Use(s.ValidateTokens())
		r.Post("/tasks/{taskID}/decline/{approvalLogin}", s.DeclineTaskHandler)
	})
	return r
}

// Run Task
// @ID RunTask
// @Security access_token
// @Security refresh_token
// @tags Работа с сервисом создания и согласования задач
// @Summary Создание задачи согласования
// @Description Создание задачи согласования с последующей отправкой
// @Param RunTask body models.RunTask true "Run Task"
// @Success 200 {object} models.Task
// @Failure 400 {object} e.ErrApiBadRequest
// @Failure 403 {object} e.ErrApiAuthFailed
// @Failure 500 {object} e.ErrApiInternalServerError
// @Router /tasks/run [post]
func (s *Server) RunTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()
	login := ctx.Value("login").(string)

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.logger.Error().Msg(err.Error())
		http.Error(w, e.JsonErrWrapper{E: err.Error()}.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	runnedTask := models.RunTask{}
	err = json.Unmarshal(data, &runnedTask)
	if err != nil {
		s.logger.Error().Msg(err.Error())
		http.Error(w, e.ErrInvalidJsonBody.Error(), http.StatusBadRequest)
		return
	}

	createdTask, err := runnedTask.CreateTask(login)
	if err != nil {
		s.logger.Error().Msg(err.Error())
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	err = s.task.RunTask(ctx, createdTask)
	if err != nil {
		s.logger.Error().Msg(err.Error())
		http.Error(w, e.JsonErrWrapper{E: err.Error()}.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(createdTask)
}

// Get Tasks List
// @ID GetTasksList
// @Security access_token
// @Security refresh_token
// @tags Работа с сервисом создания и согласования задач
// @Summary Получение списка задач
// @Description Получения списка задач пользователя
// @Success 200 {object} models.Task
// @Failure 400 {object} e.ErrApiBadRequest
// @Failure 403 {object} e.ErrApiAuthFailed
// @Failure 500 {object} e.ErrApiInternalServerError
// @Router /tasks/ [get]
func (s *Server) GetTasksListHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()

	login := ctx.Value("login").(string)

	t, err := s.task.ListTasks(ctx, login)
	if err != nil {
		s.logger.Error().Msg(err.Error())
		http.Error(w, e.JsonErrWrapper{E: err.Error()}.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(t)
}

// Update Task
// @ID UpdateTask
// @Security access_token
// @Security refresh_token
// @tags Работа с сервисом создания и согласования задач
// @Summary Обновление задачи согласования
// @Description Внесение изменений в задачу согласования в части наименования и описания задачи
// @Param taskID path string required "Task ID" Format(uuid)
// @Param UpdateTask body models.UpdateTask true "Update Task"
// @Success 200 {object} StatusUpdated
// @Failure 400 {object} e.ErrApiBadRequest
// @Failure 403 {object} e.ErrApiAuthFailed
// @Failure 404 {object} e.ErrApiNotFound
// @Failure 500 {object} e.ErrApiInternalServerError
// @Router /tasks/{taskID} [put]
func (s *Server) UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()
	login := ctx.Value("login").(string)

	id := chi.URLParam(r, "taskID")

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.logger.Error().Msg(err.Error())
		http.Error(w, e.JsonErrWrapper{E: err.Error()}.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	updateTask := models.UpdateTask{}
	err = json.Unmarshal(data, &updateTask)
	if err != nil {
		s.logger.Error().Msg(err.Error())
		http.Error(w, e.ErrInvalidJsonBody.Error(), http.StatusBadRequest)
		return
	}

	err = s.task.UpdateTask(ctx, id, login, updateTask.Name, updateTask.Text)
	if errors.Is(err, e.ErrNotFound) {
		s.logger.Error().Msg(err.Error())
		http.Error(w, e.JsonErrWrapper{E: err.Error()}.Error(), http.StatusNotFound)
		return
	} else if err != nil {
		s.logger.Error().Msg(err.Error())
		http.Error(w, e.JsonErrWrapper{E: err.Error()}.Error(), http.StatusBadRequest)
		return
	}

	log.Println(login, id)
	json.NewEncoder(w).Encode(StatusApproved{Status: "updated"})
}

// Approve Task
// @ID ApproveTask
// @Security access_token
// @Security refresh_token
// @tags Работа с сервисом создания и согласования задач
// @Summary Согласование задачи
// @Description Согласование задачи. В результате очередь согласования перейдет к следующему в списке согласующих, либо, в случае последнего этапа согласования, задача будет считаться выполненной.
// @Param taskID path string required "Task ID" Format(uuid)
// @Param approvalLogin path string required "Approval Login"
// @Success 200 {object} StatusApproved
// @Failure 400 {object} e.ErrApiBadRequest
// @Failure 403 {object} e.ErrApiAuthFailed
// @Failure 404 {object} e.ErrApiNotFound
// @Failure 500 {object} e.ErrApiInternalServerError
// @Router /tasks/{taskID}/approve/{approvalLogin} [post]
func (s *Server) ApproveTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()
	login := ctx.Value("login").(string)

	id := chi.URLParam(r, "taskID")
	approvalLogin := chi.URLParam(r, "approvalLogin")

	err := s.task.ApproveTask(ctx, login, id, approvalLogin)
	if errors.Is(err, e.ErrNotFound) {
		s.logger.Error().Msg(err.Error())
		http.Error(w, e.JsonErrWrapper{E: err.Error()}.Error(), http.StatusNotFound)
		return
	} else if err != nil {
		s.logger.Error().Msg(err.Error())
		http.Error(w, e.JsonErrWrapper{E: err.Error()}.Error(), http.StatusBadRequest)
		return
	}
	// w.Write([]byte("{\"status\": \"approved\"}"))
	json.NewEncoder(w).Encode(StatusApproved{Status: "approved"})
}

// Decline Task
// @ID DeclineTask
// @Security access_token
// @Security refresh_token
// @tags Работа с сервисом создания и согласования задач
// @Summary Отклонение задачи
// @Description Отклонение согласования задачи. В этом случае всем участникам поступит письмо с завершением задачи.
// @Param taskID path string required "Task ID" Format(uuid)
// @Param approvalLogin path string required "Approval Login"
// @Success 200 {object} StatusDeclined
// @Failure 400 {object} e.ErrApiBadRequest
// @Failure 403 {object} e.ErrApiAuthFailed
// @Failure 404 {object} e.ErrApiNotFound
// @Failure 500 {object} e.ErrApiInternalServerError
// @Router /tasks/{taskID}/decline/{approvalLogin} [post]
func (s *Server) DeclineTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()
	login := ctx.Value("login").(string)

	id := chi.URLParam(r, "taskID")
	approvalLogin := chi.URLParam(r, "approvalLogin")
	err := s.task.DeclineTask(ctx, login, id, approvalLogin)
	if errors.Is(err, e.ErrNotFound) {
		s.logger.Error().Msg(err.Error())
		http.Error(w, e.JsonErrWrapper{E: err.Error()}.Error(), http.StatusNotFound)
		return
	} else if err != nil {
		s.logger.Error().Msg(err.Error())
		http.Error(w, e.JsonErrWrapper{E: err.Error()}.Error(), http.StatusBadRequest)
		return
	}
	// w.Write([]byte("{\"status\": \"declined\"}"))
	json.NewEncoder(w).Encode(StatusDeclined{Status: "declined"})
}

// Delete Task
// @ID DeleteTask
// @Security access_token
// @Security refresh_token
// @tags Работа с сервисом создания и согласования задач
// @Summary Удаление созданной задачи
// @Description Удаление созданной задачи (доступно для автора задачи)
// @Param taskID path string required "Task ID" Format(uuid)
// @Success 200 {object} StatusDeleted
// @Failure 400 {object} e.ErrApiBadRequest
// @Failure 403 {object} e.ErrApiAuthFailed
// @Failure 404 {object} e.ErrApiNotFound
// @Failure 500 {object} e.ErrApiInternalServerError
// @Router /tasks/{taskID} [delete]
func (s *Server) DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()
	login := ctx.Value("login").(string)

	id := chi.URLParam(r, "taskID")
	err := s.task.DeleteTask(ctx, login, id)
	if errors.Is(err, e.ErrNotFound) {
		s.logger.Error().Msg(err.Error())
		http.Error(w, e.JsonErrWrapper{E: err.Error()}.Error(), http.StatusNotFound)
		return
	} else if err != nil {
		s.logger.Error().Msg(err.Error())
		http.Error(w, e.JsonErrWrapper{E: err.Error()}.Error(), http.StatusBadRequest)
		return
	}
	// w.Write([]byte("{\"status\": \"deleted\"}"))
	json.NewEncoder(w).Encode(StatusDeleted{Status: "deleted"})
}
