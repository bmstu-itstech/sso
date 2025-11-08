package main

import (
	"fmt"
	"github.com/bmstu-itstech/sso/internal/config"
	"log"
)

func main() {
	if err := config.InitConfig(); err != nil {
		log.Fatal(err)
	}
	cfg := config.GetConfig()
	fmt.Println(cfg)
	// TODO: logger

	// TODO: server

	// TODO: поднять сервер
}
