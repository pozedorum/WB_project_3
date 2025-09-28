package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	di "github.com/pozedorum/WB_project_3/task6/internal/DI"
	"github.com/pozedorum/WB_project_3/task6/pkg/config"
	"github.com/wb-go/wbf/zlog"
)

func main() {
	zlog.Init()
	cfg := config.Load()

	aplicationContainer, err := di.NewContainer(cfg)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("Failed to start container")
		return
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := aplicationContainer.Start(); err != nil {
			zlog.Logger.Fatal().Err(err).Msg("Failed to start server")
			return
		}
	}()

	<-quit
	zlog.Logger.Debug().Msg("Recived shutdown signal")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := aplicationContainer.Shutdown(ctx); err != nil {
		zlog.Logger.Fatal().Err(err).Msg("Forced shutdown")
	}
}
