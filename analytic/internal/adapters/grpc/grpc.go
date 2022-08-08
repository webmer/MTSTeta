package grpc

import (
	"context"
	"gitlab.com/g6834/team26/analytic/internal/domain/models"
	"time"

	"gitlab.com/g6834/team26/analytic/pkg/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcAuth struct {
	GrpcClient api.AuthClient
	GrpcConn   *grpc.ClientConn
}

func New(url string) (*GrpcAuth, error) {
	conn, err := grpc.Dial(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &GrpcAuth{
		GrpcClient: api.NewAuthClient(conn),
		GrpcConn:   conn,
	}, nil
}

func (GrpcAuth *GrpcAuth) Validate(tokens models.TokenPair) (*api.AuthResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	authReq := &api.AuthRequest{Service: "analytic", AccessToken: tokens.AccessToken.Value, RefreshToken: tokens.RefreshToken.Value}
	grpcResponse, err := GrpcAuth.GrpcClient.VerifyToken(ctx, authReq)
	if err != nil {
		return nil, err
	}
	return grpcResponse, nil
}

func (GrpcAuth *GrpcAuth) Stop(ctx context.Context) error {
	err := GrpcAuth.GrpcConn.Close()
	if err != nil {
		return err
	}
	return nil
}
