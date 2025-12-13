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

const port = "44044"

type Suite struct {
	*testing.B
	AuthClient ssov1.AuthClient
}

func New(b *testing.B) (context.Context, *Suite) {
	b.Helper()

	ctx, cancelCtx := context.WithTimeout(context.Background(), 5*time.Second)
	b.Cleanup(func() {
		b.Helper()
		cancelCtx()
	})

	cc, err := grpc.DialContext(context.Background(),
		net.JoinHostPort("localhost", port),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		b.Fatalf("failed to dial grpc server: %v", err)
	}
	return ctx, &Suite{
		B:          b,
		AuthClient: ssov1.NewAuthClient(cc),
	}
}
