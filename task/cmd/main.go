package main

import (
	"context"
	"os/signal"
	"syscall"

	"gitlab.com/g6834/team26/task/internal/application"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	go application.Start(ctx)
	<-ctx.Done()
	application.Stop(ctx)
}
