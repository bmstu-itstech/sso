package httpapp

import (
	"context"
	"errors"
	"fmt"
	"github.com/bmstu-itstech/sso/internal/config"
	authgrpc "github.com/bmstu-itstech/sso/internal/grpc/auth"
	http_server "github.com/bmstu-itstech/sso/internal/http"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"time"
)

type App struct {
	log    *slog.Logger
	router *gin.Engine
	srv    *http.Server
	port   int
	cfg    *config.Config
}

func New(log *slog.Logger, authServer authgrpc.AuthGrpc, cfg *config.Config) *App {
	httpSrv := http_server.New(authServer, cfg).InitRouter()
	return &App{
		log:    log,
		router: httpSrv,
		cfg:    cfg,
		port:   cfg.HTTP.Port,
	}
}

func (a *App) MustRun() {
	a.srv = &http.Server{
		Addr:    fmt.Sprintf(":%d", a.port),
		Handler: a.router,
	}

	a.log.Info("Starting server on port %d", a.port)
	if err := a.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(fmt.Sprintf("fail to start: %v", err))
	}
}

func (a *App) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	a.log.Info("Shutting down server")

	if err := a.srv.Shutdown(ctx); err != nil {
		a.log.Error("Server forced to shutdown", "error", err)
		return
	}

	a.log.Info("HTTP server stopped gracefully")
}
