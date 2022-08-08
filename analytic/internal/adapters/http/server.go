package http

import (
	"context"
	"errors"
	"fmt"
	httpSwagger "github.com/swaggo/http-swagger"
	"gitlab.com/g6834/team26/analytic/docs"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"gitlab.com/g6834/team26/analytic/internal/ports"
	"gitlab.com/g6834/team26/analytic/pkg/config"
	mv "gitlab.com/g6834/team26/analytic/pkg/middleware"
)

type Server struct {
	analytic ports.Analytic
	server   *http.Server
	l        net.Listener
	logger   *zerolog.Logger
	config   *config.Config
	port     int
}

func New(logger *zerolog.Logger, analytic ports.Analytic, config *config.Config) (*Server, error) {
	var (
		err error
		s   Server
	)
	p := fmt.Sprintf(":%s", config.Server.Port)
	s.l, err = net.Listen("tcp", p)
	if err != nil {
		logger.Error().Msgf("failed listen port %s", err)
	}
	s.config = config
	s.logger = logger
	s.analytic = analytic
	s.port = s.l.Addr().(*net.TCPAddr).Port

	s.server = &http.Server{
		Handler: s.routes(),
	}

	return &s, nil
}

func (s *Server) Port() int {
	return s.port
}

func (s *Server) Start() error {
	if err := s.server.Serve(s.l); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) routes() http.Handler {
	r := chi.NewMux()

	r.Use(mv.LoggerMiddleware(s.logger))
	r.Use(mv.RecovererMiddleware(s.logger))
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/healths", s.healthsHandler)
	r.Mount("/analytic/v1", s.analyticHandlers())

	r.Post("/toggle-prof", s.toggleDebugHandler)
	r.Mount("/debug", s.debugHandlers())

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http:%v%v/swagger/doc.json", docs.SwaggerInfo.Host, docs.SwaggerInfo.BasePath))))

	return r
}

func (s *Server) healthsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
