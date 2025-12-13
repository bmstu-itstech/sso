package main

import (
	"context"
	"github.com/bmstu-itstech/sso/client_sso_by_scriptum"
	"github.com/bmstu-itstech/sso/internal/logs"
	"log/slog"
)

func main() {
	log := logs.NewLogger("example")
	sso, stop := client_sso_by_scriptum.MustNewSSOClient(
		client_sso_by_scriptum.Config{
			AppId: 1,
			Host:  "localhost",
			Port:  "44044"},
		log)

	defer func() {
		if err := stop(); err != nil {
			log.Error("failed to close sso client", slog.String("error", err.Error()))
		}
	}()

	ctx := context.Background()
	isAdmin, err := sso.IsAdmin(ctx, 8268673725144469736)

	if err != nil {
		log.Error("failed to check admin status", slog.String("error", err.Error()))
		return
	}
	log.Info("No error checking admin status", slog.Bool("is_admin", isAdmin))
}
