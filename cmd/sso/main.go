package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"sso/internal/app"
	"sso/internal/config"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

func main() {
	// TODO: config init
	cfg := config.MustLoad()

	//TODO: logger init
	log := setupLogger(cfg.Env)
	log.Info("starting application",
		slog.String("env", cfg.Env),
		slog.Any("cfg", cfg),
		slog.Int("port", cfg.GRPC.Port),
	)

	application := app.New(log, cfg.GRPC.Port, cfg.StoragePath, cfg.TokenTTL)

	go application.GRPCSrv.MustRun()

	//TODO: app init

	//TODO: gRPC-server start

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	sign := <-stop
	log.Info("caught signal; shutting down... ", sign.String())
	application.GRPCSrv.Stop()
	log.Info("application stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return log
}
