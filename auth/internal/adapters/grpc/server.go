package grpc

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"gitlab.com/g6834/team26/auth/internal/ports"
	"gitlab.com/g6834/team26/auth/pkg/api"
	"gitlab.com/g6834/team26/auth/pkg/config"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"net"
)

type Server struct {
	auth   ports.Auth
	server *grpc.Server
	l      net.Listener
	logger *zerolog.Logger
	config *config.Config
}

func New(logger *zerolog.Logger, auth ports.Auth, config *config.Config) (*Server, error) {
	var (
		err error
		s   Server
	)
	p := fmt.Sprintf(":%s", config.Server.GRPCPort)
	s.l, err = net.Listen("tcp", p)
	if err != nil {
		logger.Error().Msgf("failed listen port %s", err)
	}
	s.config = config
	s.logger = logger
	s.auth = auth

	s.server = grpc.NewServer()
	api.RegisterAuthServer(s.server, &AuthServer{authS: auth})

	return &s, nil
}

func (s *Server) Start() error {
	if err := s.server.Serve(s.l); !errors.Is(err, errors.New("grpc: Server closed")) {
		return err
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.server.GracefulStop()
	return nil
}
