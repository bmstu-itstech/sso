package app

import (
	"github.com/bmstu-itstech/sso/internal/services/jwt"
	"log"
	"log/slog"

	grpcapp "github.com/bmstu-itstech/sso/internal/app/grpc"
	"github.com/bmstu-itstech/sso/internal/config"
	"github.com/bmstu-itstech/sso/internal/repository/postgres"
	"github.com/bmstu-itstech/sso/internal/services"
)

type App struct {
	GRPCSrv *grpcapp.App
	cfg     *config.Config
}

func New(logger *slog.Logger, cfg config.Config) *App {
	repos, err := postgres.NewPostgresDB(cfg.Postgres)
	jwtService := jwt.NewServiceJwt(cfg.JWT.TokenTTL)
	if err != nil {
		log.Fatal("db no connect", slog.String("error", err.Error()))
	}
	authService := services.New(logger, &repos, &repos, &repos, jwtService, &cfg)
	grpcSrv := grpcapp.New(logger, authService, cfg)
	return &App{
		GRPCSrv: grpcSrv,
		cfg:     &cfg,
	}
}
