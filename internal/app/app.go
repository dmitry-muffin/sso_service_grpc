package app

import (
	"log/slog"
	grpcapp "sso/internal/app/grpc"
	httpapp "sso/internal/app/http"
	authhttp "sso/internal/http/auth"
	"sso/internal/services/auth"
	"sso/internal/storage/postgres"
	"time"
)

type App struct {
	GRPCSrv *grpcapp.App
	HTTPSrv *httpapp.Srv
}

func New(
	log *slog.Logger,
	grpcPort int,
	dbHost string,
	dbPort int,
	dbUser string,
	dbPassword string,
	dbName string,
	httpAddr int,
	httpTimeout time.Duration,
	httpIdle time.Duration,
	tokenTTL time.Duration,
) *App {

	dsn := postgres.DSN(
		dbHost,
		dbPort,
		dbUser,
		dbPassword,
		dbName,
		"disable", // sslmode
	)
	//TODO: init storage

	storage, err := postgres.New(dsn)
	if err != nil {
		panic(err)
	}
	//TODO: init auth service

	authService := auth.New(log, storage, storage, storage, tokenTTL)

	grpcApp := grpcapp.New(log, authService, grpcPort)

	httpHandlers := authhttp.NewHandler(storage, log, tokenTTL)
	httpServ := httpapp.New(log, httpHandlers, httpAddr)
	return &App{
		GRPCSrv: grpcApp,
		HTTPSrv: httpServ,
	}
}

func (a *App) Stop() {
	a.GRPCSrv.Stop()
	a.HTTPSrv.Stop()
}
