package suite

import (
	"context"
	ssov1 "github.com/BOBAvov/protos_sso/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
	"time"
)

const port = "8080"

type Suite struct {
	*testing.T
	AuthClient ssov1.AuthClient
}

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()
	t.Parallel()

	ctx, cancelCtx := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(func() {
		t.Helper()
		cancelCtx()
	})

	cc, err := grpc.DialContext(context.Background(),
		net.JoinHostPort("localhost", port),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to dial grpc server: %v", err)
	}
	return ctx, &Suite{
		T:          t,
		AuthClient: ssov1.NewAuthClient(cc),
	}
}
