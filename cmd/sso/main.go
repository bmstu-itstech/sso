package main

import (
	"fmt"
	"github.com/bmstu-itstech/sso/internal/app"
	"github.com/bmstu-itstech/sso/internal/config"
	"github.com/bmstu-itstech/sso/internal/logs"
	"log"
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
	logger.Info("Hello world")

	application := app.New(logger, cfg.GRPC.Port)

	application.GRPCSrv.MustRun()

	// TODO: server

	// TODO: поднять сервер
}
