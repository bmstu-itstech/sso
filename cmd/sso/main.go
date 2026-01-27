package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/bmstu-itstech/sso/internal/app"
	"github.com/bmstu-itstech/sso/internal/config"
	"github.com/bmstu-itstech/sso/internal/logs"
)

func main() {
	if err := config.InitConfig(); err != nil {
		log.Fatal(err)
	}

	cfg := config.GetConfig()

	logger := logs.NewLogger(cfg.ENV)
	logger.Info("Server started")

	application := app.New(logger, *cfg)

	go application.GRPCSrv.MustRun()
	go application.HTTPrv.MustRun()

	logger.Info(fmt.Sprintf("docx to: http://localhost:%s/swagger/index.html", strconv.Itoa(cfg.HTTP.Port)))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	sign := <-stop

	logger.Info("Server stopped by signal", slog.String("signal", sign.String()))

	application.GRPCSrv.Stop()
	application.HTTPrv.Stop()

	logger.Info("GRPC Server stopped")
	logger.Info("HTTP Server stopped")

}
