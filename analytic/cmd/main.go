package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"gitlab.com/g6834/team26/analytic/internal/application"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer cancel()

	app := application.New(ctx)
	go app.Start()
	<-ctx.Done()
	app.Stop()
}
