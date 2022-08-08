package application

import (
	"context"
	"github.com/rs/zerolog"
	"gitlab.com/g6834/team26/auth/internal/adapters/grpc"
	"gitlab.com/g6834/team26/auth/internal/adapters/http"
	"gitlab.com/g6834/team26/auth/internal/adapters/mongo"
	"gitlab.com/g6834/team26/auth/internal/domain/auth"
	"gitlab.com/g6834/team26/auth/pkg/config"
	"gitlab.com/g6834/team26/auth/pkg/logger"
	"golang.org/x/sync/errgroup"
	"os"
)

type App struct {
	s  *http.Server
	g  *grpc.Server
	db *mongo.Database
	l  *zerolog.Logger
}

func New(ctx context.Context) *App {
	l := logger.New()

	c, err := config.New()
	if err != nil {
		l.Error().Msgf("error parsing env: %s", err)
		os.Exit(1)
	}

	db, err := mongo.New(ctx, c)
	if err != nil {
		l.Error().Msgf("db init failed: %s", err)
		os.Exit(1)
	}
	authS := auth.New(db, c)

	s, err := http.New(l, authS, c)
	if err != nil {
		l.Error().Msgf("http server creating failed: %s", err)
	}
	g, err := grpc.New(l, authS, c)
	if err != nil {
		l.Error().Msgf("grpc server creating failed: %s", err)
	}

	return &App{
		s:  s,
		g:  g,
		db: db,
		l:  l,
	}
}

func (a *App) Start() {
	var eg errgroup.Group
	eg.Go(func() error {
		return a.s.Start()
	})
	eg.Go(func() error {
		return a.g.Start()
	})

	a.l.Info().Msg("app is started")
	err := eg.Wait()
	if err != nil {
		a.l.Error().Msgf("server start failed %s", err)
	}
}

func (a *App) Stop() {
	_ = a.s.Stop(context.Background())
	_ = a.g.Stop(context.Background())
	_ = a.db.Disconnect(context.Background())
	a.l.Info().Msg("app has stopped")

}
