package suite

import (
	"context"
	"fmt"
	ssov1 "github.com/BOBAvov/protos_sso/gen/go/sso"
	"github.com/bmstu-itstech/sso/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"strconv"
	"testing"
	"time"
)

type Suite struct {
	*testing.T
	Cfg        *config.Config
	AuthClient ssov1.AuthClient
}

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()
	t.Parallel()

	cfg := config.GetConfig()
	ctx, cancelCtx := context.WithTimeout(context.Background(), 5*time.Second)
	fmt.Println("localhost:" + strconv.Itoa(cfg.GRPC.Port))
	t.Cleanup(func() {
		t.Helper()
		cancelCtx()
	})

	cc, err := grpc.DialContext(context.Background(),
		net.JoinHostPort("localhost", "8080"),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to dial grpc server: %v", err)
	}
	return ctx, &Suite{
		T:          t,
		Cfg:        cfg,
		AuthClient: ssov1.NewAuthClient(cc),
	}
}
