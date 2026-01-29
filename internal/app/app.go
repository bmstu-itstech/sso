package app

import (
	"log"
	"log/slog"

	httpapp "github.com/bmstu-itstech/sso/internal/app/http"
	"github.com/bmstu-itstech/sso/internal/repository/cache"
	"github.com/bmstu-itstech/sso/internal/repository/redis"
	"github.com/bmstu-itstech/sso/internal/services/jwt"

	grpcapp "github.com/bmstu-itstech/sso/internal/app/grpc"
	"github.com/bmstu-itstech/sso/internal/config"
	"github.com/bmstu-itstech/sso/internal/repository/postgres"
	"github.com/bmstu-itstech/sso/internal/services"
)

type App struct {
	GRPCSrv *grpcapp.App
	cfg     *config.Config
	HTTPrv  *httpapp.App
}

func New(logger *slog.Logger, cfg config.Config) *App {
	repos, err := postgres.NewPostgresDB(cfg)
	cacheApp, err := cache.NewAppCache(cfg)
	jwtService := jwt.NewServiceJwt(cfg.JWT.TokenTTL)
	eventService := redis.New(cfg)
	if err != nil {
		log.Fatal("db no connect", slog.String("error", err.Error()))
	}
	authService := services.New(logger, &repos, &repos, cacheApp, jwtService, eventService, &cfg)
	grpcSrv := grpcapp.New(logger, authService, cfg)
	httpSrv := httpapp.New(logger, authService, &cfg)
	return &App{
		GRPCSrv: grpcSrv,
		HTTPrv:  httpSrv,
		cfg:     &cfg,
	}
}
