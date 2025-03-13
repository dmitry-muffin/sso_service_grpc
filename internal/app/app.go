package app

import (
	"log/slog"
	grpcapp "sso/internal/app/grpc"
	"time"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(log *slog.Logger, grpcPort int, storagePath string, tokenTTL time.Duration) *App {

	//TODO: init storage

	//TODO: init auth service

	grpcApp := grpcapp.New(log, grpcPort)

	return &App{
		GRPCSrv: grpcApp,
	}
}
