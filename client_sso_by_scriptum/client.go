package client_sso_by_scriptum

import (
	"context"
	ssov1 "github.com/BOBAvov/protos_sso/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"log/slog"
	"net"
	"strconv"
)

type Config struct {
	Host  string
	Port  string
	AppId int32
}

type SSO struct {
	conn  *grpc.ClientConn
	api   ssov1.AuthClient
	l     *slog.Logger
	AppId int32
}

func MustNewSSOClient(config Config, l *slog.Logger) (*SSO, func() error) {
	sso, closeFn, err := NewSSOClient(config, l)
	if err != nil {
		panic("Error creating SSO client: " + err.Error())
	}
	return sso, closeFn
}

func NewSSOClient(config Config, l *slog.Logger) (*SSO, func() error, error) {
	addr := net.JoinHostPort(config.Host, config.Port)
	cc, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, func() error { return nil }, err
	}

	closeFn := cc.Close

	return &SSO{
		api:   ssov1.NewAuthClient(cc),
		conn:  cc,
		l:     l,
		AppId: config.AppId,
	}, closeFn, nil
}

// IsAdmin checks if the user with the given uid has admin privileges. uid is int64!!!
func (s *SSO) IsAdmin(ctx context.Context, uid int64) (bool, error) {
	const op = "client_sso_by_scriptum.IsAdmin"

	l := s.l.With(
		slog.String("op", op),
		slog.Int64("uid", uid),
	)

	l.Debug("Checking admin status")
	// Добавляем метаданные с AppId
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(
		"x-user-id", strconv.FormatInt(int64(s.AppId), 10),
	))
	// Вызываем метод IsAdmin на сервере SSO
	resp, err := s.api.IsAdmin(ctx, &ssov1.IsAdminRequest{
		UserId: uid,
	})
	if err != nil {
		l.Error("Failed to check admin status: ", err.Error())
		return false, err
	}

	l.Debug("Admin status: ", resp.IsAdmin)

	return resp.IsAdmin, nil
}
