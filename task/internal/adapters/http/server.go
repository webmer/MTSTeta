package http

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	httpSwagger "github.com/swaggo/http-swagger"
	"gitlab.com/g6834/team26/task/docs"
	"gitlab.com/g6834/team26/task/internal/ports"
	"gitlab.com/g6834/team26/task/pkg/config"
	httpMiddleware "gitlab.com/g6834/team26/task/pkg/middleware"
)

type Server struct {
	task     ports.Task
	server   *http.Server
	logger   *zerolog.Logger
	listener net.Listener
	config   *config.Config
	port     int
}

func New(l *zerolog.Logger, task ports.Task, config *config.Config) (*Server, error) {
	var (
		err error
		s   Server
	)
	port := fmt.Sprintf(":%s", config.Server.Port)
	s.listener, err = net.Listen("tcp", port)
	if err != nil {
		log.Fatal("Failed listen port", err)
	}
	s.config = config
	s.task = task
	s.logger = l
	s.port = s.listener.Addr().(*net.TCPAddr).Port

	s.server = &http.Server{
		Handler: s.routes(),
	}

	return &s, nil
}

func (s *Server) Port() int {
	return s.port
}

func (s *Server) Start(ctx context.Context) error {
	// ctx := context.Background()
	go s.task.StartMessageSender(ctx)
	go s.task.StartEmailSender(ctx)

	if err := s.server.Serve(s.listener); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) routes() http.Handler {
	r := chi.NewRouter()
	r.Use(httpMiddleware.LoggerMiddleware(s.logger))
	r.Use(httpMiddleware.RecovererMiddleware(s.logger))
	r.Use(middleware.Timeout(60 * time.Second))

	r.Mount("/task/v1", s.taskHandlers())

	r.Post("/toggle-prof", s.toggleDebugHandler)
	r.Mount("/debug", s.debugHandlers())

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http:%v%v/swagger/doc.json", docs.SwaggerInfo.Host, docs.SwaggerInfo.BasePath))))

	return r
}
