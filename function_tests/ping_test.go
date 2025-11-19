package function_tests

import (
	"github.com/bmstu-itstech/sso/function_tests/suite"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
	"testing"
)

func TestPing(t *testing.T) {
	ctx, st := suite.New(t)
	_, err := st.AuthClient.Ping(ctx, &emptypb.Empty{})
	require.NoError(t, err)
}
