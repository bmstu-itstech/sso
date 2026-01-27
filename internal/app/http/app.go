package httpapp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/bmstu-itstech/sso/internal/config"
	http_server "github.com/bmstu-itstech/sso/internal/http"
	"github.com/gin-gonic/gin"
)

type App struct {
	log    *slog.Logger
	router *gin.Engine
	srv    *http.Server
	port   int
	cfg    *config.Config
}

func New(log *slog.Logger, authServer http_server.Auth, cfg *config.Config) *App {
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

	a.log.Info("http server starting", slog.Int("port", a.port))
	if err := a.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(fmt.Sprintf("fail to start: %v", err))
	}
}

// eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhcHBfaWQiOjEsImV4cCI6MTc2OTU0NjM4MCwiaXNfYWRtaW4iOnsiVWlkIjo0OTgzNzY3MDA1NDQ4MTYwODU3LCJBcHBJZCI6MSwiU2VjcmV0IjoidGVzdC1zZWNyZXQiLCJJc0FkbWluIjpmYWxzZX0sInVpZCI6IjQ5ODM3NjcwMDU0NDgxNjA4NTcifQ.5wrP_2bg7m72eLlsLlRtr4yTNQsfuhO227YJPESUTdo

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
