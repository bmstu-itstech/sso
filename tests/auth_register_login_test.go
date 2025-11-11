package tests

import (
	ssov1 "github.com/BOBAvov/protos_sso/gen/go/sso"
	"github.com/bmstu-itstech/sso/internal/lib"
	"github.com/bmstu-itstech/sso/tests/suite"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

const (
	emptyAppId = 0
	appId      = 1
	appSecret  = "test-secret"

	passDefaultLen = 10
)

func TestRegisterLogin_Login_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	login := gofakeit.Username()
	fullName := gofakeit.Name()
	password := randomFakePassword()

	t.Log(email, login, fullName, password)

	respRegister, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Login:    login,
		Password: password,
		Email:    email,
		FullName: fullName,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respRegister.GetUserId())

	t.Log(respRegister.GetUserId())

	respLogin, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		AppId:    appId,
		Login:    login,
		Password: password,
	})

	require.NoError(t, err)

	token := respLogin.GetToken()
	require.NotEmpty(t, token)

	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})
	require.NoError(t, err)

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	require.True(t, ok)

	t.Log(respRegister.GetUserId(), claims["uid"])
	assert.Equal(t, strconv.FormatInt(respRegister.GetUserId(), 10), claims["uid"].(string))
	assert.Equal(t, login, claims["login"].(string))
	assert.Equal(t, email, claims["email"].(string))
	assert.Equal(t, appId, int(claims["app_id"].(float64)))

}

func randomFakePassword() string {
	return gofakeit.Password(true, true, true, true, false, passDefaultLen)
}

func TestIsAdminIsFakeId(t *testing.T) {
	ctx, st := suite.New(t)
	id := lib.RandoInt64()
	t.Log(id)
	resp, err := st.AuthClient.IsAdmin(ctx, &ssov1.IsAdminRequest{UserId: id})
	require.NoError(t, err)
	assert.False(t, resp.IsAdmin)
}
func TestIsAdminIsNotFakeId(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	login := gofakeit.Username()
	fullName := gofakeit.Name()
	password := randomFakePassword()

	t.Log(email, login, fullName, password)

	respRegister, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Login:    login,
		Password: password,
		Email:    email,
		FullName: fullName,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respRegister.GetUserId())

	t.Log(respRegister.GetUserId())

	respLogin, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		AppId:    appId,
		Login:    login,
		Password: password,
	})

	require.NoError(t, err)

	token := respLogin.GetToken()
	require.NotEmpty(t, token)

	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})
	require.NoError(t, err)

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	require.True(t, ok)

	t.Log(respRegister.GetUserId(), claims["uid"])
	assert.Equal(t, strconv.FormatInt(respRegister.GetUserId(), 10), claims["uid"].(string))
	assert.Equal(t, login, claims["login"].(string))
	assert.Equal(t, email, claims["email"].(string))
	assert.Equal(t, appId, int(claims["app_id"].(float64)))

	respIsAdmin, err := st.AuthClient.IsAdmin(ctx, &ssov1.IsAdminRequest{
		UserId: respRegister.GetUserId(),
	})
	require.NoError(t, err)
	assert.False(t, respIsAdmin.IsAdmin)
}
