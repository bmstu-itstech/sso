package http_test

import (
	"net/http"
	"testing"

	"github.com/bmstu-itstech/sso/tests/http_test/suite"
	"github.com/stretchr/testify/require"
)

func TestPing_HappyPath(t *testing.T) {
	st := suite.New(t)

	rr := st.DoJSON(http.MethodGet, "/api/v1/ping", nil, nil)
	require.Equal(t, http.StatusOK, rr.Code)
}

func TestPing_WrongMethod(t *testing.T) {
	st := suite.New(t)

	rr := st.DoJSON(http.MethodPost, "/api/v1/ping", nil, nil)
	require.NotEqual(t, http.StatusOK, rr.Code)
}
