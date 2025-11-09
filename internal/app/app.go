package app

import (
	grpcapp "github.com/bmstu-itstech/sso/internal/app/grpc"
	"github.com/bmstu-itstech/sso/internal/config"
	authgrpc "github.com/bmstu-itstech/sso/internal/grpc/auth"
	"github.com/bmstu-itstech/sso/internal/repository"
	"github.com/bmstu-itstech/sso/internal/services/auth"
	"log"
	"log/slog"
)

type App struct {
	GRPCSrv *grpcapp.App
}

// New TODO: погумай где инициализировать хранилища, сервисы и прочее
func New(logger *slog.Logger, cfg config.Config) *App {
	repos, err := repository.NewPostgresDB(cfg.Postgres)
	if err != nil {
		log.Fatal("db no connect", slog.String("error", err.Error()))
	}
	authService := auth.New(logger, &repos, &repos, &repos, cfg.JWT.TokenTTL)
	grpcAuth := &authgrpc.Auth{
		Login:           authService.Login,
		RegisterNewUser: authService.RegisterNewUser,
		IsAdmin:         authService.IsAdmin,
	}
	grpcSrv := grpcapp.New(logger, grpcAuth, cfg.GRPC.Port)
	return &App{
		GRPCSrv: grpcSrv,
	}
}
