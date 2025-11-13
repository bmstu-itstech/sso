package app

import (
	grpcapp "github.com/bmstu-itstech/sso/internal/app/grpc"
	"github.com/bmstu-itstech/sso/internal/config"
	authgrpc "github.com/bmstu-itstech/sso/internal/grpc/auth"
	"github.com/bmstu-itstech/sso/internal/repository/postgres"
	"github.com/bmstu-itstech/sso/internal/services"
	"log"
	"log/slog"
)

type App struct {
	GRPCSrv *grpcapp.App
	cfg     *config.Config
}

func New(logger *slog.Logger, cfg config.Config) *App {
	repos, err := postgres.NewPostgresDB(cfg.Postgres)
	if err != nil {
		log.Fatal("db no connect", slog.String("error", err.Error()))
	}
	authService := services.New(logger, &repos, &repos, &repos, &cfg)
	grpcAuth := &authgrpc.Auth{
		Login:           authService.Login,
		RegisterNewUser: authService.RegisterNewUser,
		IsAdmin:         authService.IsAdmin,
		UserInfo:        authService.GetUserInfo,
		DeleteUser:      authService.DeleteUser,
		UpdatePassword:  authService.UpdatePassword,
		UsersAll:        authService.GetAllUsers,
	}
	grpcSrv := grpcapp.New(logger, grpcAuth, cfg)
	return &App{
		GRPCSrv: grpcSrv,
		cfg:     &cfg,
	}
}
