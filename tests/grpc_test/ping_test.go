package grpc_test

import (
	"github.com/bmstu-itstech/sso/tests/grpc_test/suite"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
	"testing"
)

func TestPing(t *testing.T) {
	ctx, st := suite.New(t)
	_, err := st.AuthClient.Ping(ctx, &emptypb.Empty{})
	require.NoError(t, err)
}
