package grpcapp

import (
	"fmt"
	"github.com/bmstu-itstech/sso/internal/grpc/middleware"
	"log/slog"
	"net"

	"google.golang.org/grpc"

	"github.com/bmstu-itstech/sso/internal/config"
	authgrpc "github.com/bmstu-itstech/sso/internal/grpc/auth"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
	cfg        *config.Config
}

func New(log *slog.Logger, authServer authgrpc.AuthGrpc, cfg config.Config) *App {
	authInterceptor := middleware.LoggerInterceptor(log)
	gRPCServer := grpc.NewServer(grpc.UnaryInterceptor(authInterceptor))

	authgrpc.RegisterServer(gRPCServer, authServer, &cfg)
	return &App{
		cfg:        &cfg,
		log:        log,
		gRPCServer: gRPCServer,
		port:       cfg.GRPC.Port,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}
func (a *App) Run() error {
	const op = "grpcapp.App.Run"
	log := a.log.With(slog.String("op", op), slog.Int("port", a.port))

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		log.Error("failed to listen", slog.String("error", err.Error()))
		return fmt.Errorf("%w: %s", err, op)
	}

	log.Info("grpc server started", slog.String("addr", l.Addr().String()))
	if err := a.gRPCServer.Serve(l); err != nil {
		log.Error("failed to serve grpc", slog.String("error", err.Error()))
		return fmt.Errorf("%w: %s", err, op)
	}

	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.App.Stop"
	log := a.log.With(slog.String("op", op))

	a.gRPCServer.GracefulStop()
	log.Info("grpc server stopped")
}
