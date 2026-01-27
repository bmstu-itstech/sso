package http_integration

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/bmstu-itstech/sso/internal/domain/models"
	"github.com/bmstu-itstech/sso/tests/http_integration_test/suite"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterLogin_HappyPath(t *testing.T) {
	st := suite.New(t)

	u := st.NewRandomUser()
	u = st.RegisterUser(u)
	u = st.LoginUser(u, appID)

	// проверяем JWT-claims аналогично grpc_test
	tokenParsed, err := jwt.Parse(u.Token, func(token *jwt.Token) (any, error) {
		return []byte(jwtSecret), nil
	})
	require.NoError(t, err)

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	require.True(t, ok)

	assert.Equal(t, strconv.FormatInt(u.UserID, 10), claims["uid"].(string))
	assert.Equal(t, float64(appID), claims["app_id"].(float64))
}

func TestRegister_ValidationError(t *testing.T) {
	st := suite.New(t)

	u := st.NewRandomUser()

	resp, body := st.DoJSON(http.MethodPost, "/api/v1/register", models.RegisterRequest{
		Login:    u.Login,
		Password: "short", // min=8
		Email:    "not-an-email",
		FullName: "",
	}, nil)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.NotEmpty(t, string(body))
}
