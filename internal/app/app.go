package app

import (
	grpcapp "github.com/bmstu-itstech/sso/internal/app/grpc"
	"log/slog"
)

type App struct {
	GRPCSrv *grpcapp.App
}

// New TODO: погумай где инициализировать хранилища, сервисы и прочее
func New(log *slog.Logger, port int) *App {
	grpcSrv := grpcapp.New(log, port)
	return &App{
		GRPCSrv: grpcSrv,
	}
}
