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
	//config init
	cfg := config.MustLoad()

	//logger init
	log := setupLogger(cfg.Env)
	log.Info("starting application",
		slog.String("env", cfg.Env),
		slog.Any("cfg", cfg),
		slog.Int("port", cfg.GRPC.Port),
	)

	// start gRPC server
	application := app.New(log,
		cfg.GRPC.Port,
		cfg.PgDb.Host,
		cfg.PgDb.Port,
		cfg.PgDb.Username,
		cfg.PgDb.Password,
		cfg.PgDb.Database,
		cfg.HTTPConf.Address,
		cfg.HTTPConf.Timeout,
		cfg.HTTPConf.IdleTimeout,
		cfg.TokenTTL)

	go func() {
		if err := application.GRPCSrv.Run(); err != nil {
			log.Error("gRPC server failed", err)
		}
	}()

	if err := application.HTTPSrv.Run(); err != nil {
		log.Error("HTTP server failed", err)
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	sign := <-quit
	log.Info("received shutdown signal", slog.String("signal", sign.String()))
	/*application.GRPCSrv.Stop()
	application.HTTPSrv.Shutdown(context.Background())
	log.Info("application stopped")*/
	application.Stop()

	//go application.GRPCSrv.MustRun()
	//
	//// Graceful shuttdown
	//stop := make(chan os.Signal, 1)
	//signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	//sign := <-stop
	//log.Info("caught signal; shutting down... ", sign.String())
	//application.GRPCSrv.Stop()
	//log.Info("application stopped")

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
