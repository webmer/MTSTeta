package application

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"gitlab.com/g6834/team26/analytic/internal/adapters/grpc"
	"gitlab.com/g6834/team26/analytic/internal/adapters/http"
	"gitlab.com/g6834/team26/analytic/internal/adapters/kafka"
	"gitlab.com/g6834/team26/analytic/internal/adapters/postgres"
	"gitlab.com/g6834/team26/analytic/internal/domain/analytic"
	"gitlab.com/g6834/team26/analytic/pkg/config"
	"gitlab.com/g6834/team26/analytic/pkg/logger"
	"golang.org/x/sync/errgroup"
	"os"
)

type App struct {
	s             *http.Server
	grpcAuth      *grpc.GrpcAuth
	kafkaConsumer *kafka.Consumer
	db            *postgres.PostgresDatabase
	l             *zerolog.Logger
}

func New(ctx context.Context) *App {
	l := logger.New()

	c, err := config.New()
	if err != nil {
		l.Error().Msgf("error parsing env: %s", err)
		os.Exit(1)
	}

	db, err := postgres.New(ctx, c.Server.AuthorizationDBConnectionString)
	if err != nil {
		l.Error().Msgf("db init failed: %s", err)
		os.Exit(1)
	}

	p := fmt.Sprintf(":%s", c.Server.GRPCPort)
	grpcAuth, err := grpc.New(p)
	if err != nil {
		l.Error().Msgf("grpc client init failed: %s", err)
		os.Exit(1)
	}

	analyticS := analytic.New(db, grpcAuth)

	s, err := http.New(l, analyticS, c)
	if err != nil {
		l.Error().Msgf("http server creating failed: %s", err)
		os.Exit(1)
	}

	kc, err := kafka.New(l, analyticS, c)
	if err != nil {
		l.Error().Msgf("kafka analytic client init failed: %s", err)
		os.Exit(1)
	}

	return &App{
		s:             s,
		grpcAuth:      grpcAuth,
		kafkaConsumer: kc,
		db:            db,
		l:             l,
	}
}

func (a *App) Start() {
	var eg errgroup.Group
	eg.Go(func() error {
		return a.s.Start()
	})
	eg.Go(func() error {
		return a.kafkaConsumer.StartRead()
	})

	a.l.Info().Msg("app is started")
	err := eg.Wait()
	if err != nil {
		a.l.Error().Msgf("server start failed %s", err)
	}
}

func (a *App) Stop() {
	_ = a.s.Stop(context.Background())
	_ = a.grpcAuth.Stop(context.Background())
	_ = a.kafkaConsumer.StopRead(context.Background())
	_ = a.db.Stop(context.Background())
	a.l.Info().Msg("app has stopped")
}
