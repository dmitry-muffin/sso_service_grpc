package app

import (
	"log/slog"
	grpcapp "sso/internal/app/grpc"
	"sso/internal/services/auth"
	"sso/storage"
	"time"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
	dbHost string,
	dbPort int,
	dbUser string,
	dbPassword string,
	dbName string,
	tokenTTL time.Duration,
) *App {

	dsn := storage.DSN(
		dbHost,
		dbPort,
		dbUser,
		dbPassword,
		dbName,
		"disable", // sslmode
	)
	//TODO: init storage

	storage, err := storage.New(dsn)
	if err != nil {
		panic(err)
	}
	//TODO: init auth service

	authService := auth.New(log, storage, storage, storage, tokenTTL)

	grpcApp := grpcapp.New(log, authService, grpcPort)

	return &App{
		GRPCSrv: grpcApp,
	}
}
