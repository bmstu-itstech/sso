package http_integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/bmstu-itstech/sso/internal/domain/models"
	"github.com/bmstu-itstech/sso/tests/http_integration_test/suite"
	"github.com/stretchr/testify/require"
)

// Аналог grpc_test/TestIsAdminIsNot, но через HTTP.
func TestIsAdmin_IsNot_HappyPath(t *testing.T) {
	st := suite.New(t)

	u := st.NewRandomUser()
	u = st.RegisterUser(u)
	u = st.LoginUser(u, appID)

	headers := map[string]string{
		"Authorization": "Bearer " + u.Token,
	}

	resp, body := st.DoJSON(http.MethodGet, "/api/v1/user/is_admin/"+fmt.Sprintf("%d", u.UserID), nil, headers)
	require.Equal(t, http.StatusOK, resp.StatusCode, "body=%s", string(body))

	var out models.IsAdminResponse
	require.NoError(t, json.Unmarshal(body, &out))
	require.False(t, out.IsAdmin)
}
