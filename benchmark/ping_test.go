package benchmark

import (
	"github.com/bmstu-itstech/sso/benchmark/suite"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
	"testing"
)

func BenchmarkPing(b *testing.B) {
	ctx, st := suite.New(b)
	_, err := st.AuthClient.Ping(ctx, &emptypb.Empty{})
	require.NoError(b, err)
}
