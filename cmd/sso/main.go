package main

import (
	"fmt"
	"github.com/bmstu-itstech/sso/internal/app"
	"github.com/bmstu-itstech/sso/internal/config"
	"github.com/bmstu-itstech/sso/internal/logs"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := config.InitConfig(); err != nil {
		log.Fatal(err)
	}
	cfg := config.GetConfig()
	if cfg.ENV == "local" {
		fmt.Println(cfg)
	}
	logger := logs.NewLogger(cfg.ENV)
	logger.Info("Server started")

	application := app.New(logger, *cfg)

	go application.GRPCSrv.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	sign := <-stop

	logger.Info("Server stopped by signal", slog.String("signal", sign.String()))

	application.GRPCSrv.Stop()

	logger.Info("Server stopped")
}
