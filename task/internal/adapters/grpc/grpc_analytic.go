package grpc

import (
	"context"
	"time"

	"gitlab.com/g6834/team26/task/pkg/api"
	"gitlab.com/g6834/team26/task/pkg/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcAnalytic struct {
	GrpcClient api.AnalyticClient
	GrpcConn   *grpc.ClientConn
}

func NewAnalytic(url string) (*GrpcAnalytic, error) {
	conn, err := grpc.Dial(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	// defer conn.Close()
	return &GrpcAnalytic{
		GrpcClient: api.NewAnalyticClient(conn),
		GrpcConn:   conn,
	}, nil
}

func (GrpcAnalytic *GrpcAnalytic) ActionTask(ctx context.Context, u, t, v string) error {
	_, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	actionTaskReq := &api.MessageRequest{
		UUID:        u,
		UUIDMessage: uuid.GenUUID(),
		Timestamp:   time.Now().Unix(),
		Type:        t,
		Value:       v,
	}
	// log.Println(addTaskReq)
	_, err := GrpcAnalytic.GrpcClient.ActionTask(ctx, actionTaskReq)
	if err != nil {
		return err
	}
	return nil
}

func (GrpcAnalytic *GrpcAnalytic) StopAnalytic(ctx context.Context) error {
	err := GrpcAnalytic.GrpcConn.Close()
	if err != nil {
		return err
	}
	return nil
}
