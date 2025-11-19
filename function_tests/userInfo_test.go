package function_tests

import (
	"strconv"
	"testing"

	ssov1 "github.com/BOBAvov/protos_sso/gen/go/sso"
	"github.com/bmstu-itstech/sso/function_tests/suite"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

// пользователь
func TestUserInfo_UserSelf(t *testing.T) {
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
		AppId:    ssoId,
		Login:    login,
		Password: password,
	})
	require.NoError(t, err)

	token := respLogin.GetToken()
	require.NotEmpty(t, token)

	// Создаем контекст с токеном для авторизованного запроса
	ctxWithToken := metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "Bearer "+token))

	respUserInfo, err := st.AuthClient.UserInfo(ctxWithToken, &ssov1.UserInfoRequest{
		UserId: respRegister.GetUserId(),
	})
	require.NoError(t, err)
	assert.Equal(t, respRegister.GetUserId(), respUserInfo.GetUserId())
	assert.Equal(t, login, respUserInfo.GetLogin())
	assert.Equal(t, email, respUserInfo.GetEmail())
	assert.Equal(t, fullName, respUserInfo.GetFullName())
	assert.False(t, respUserInfo.GetIsAdmin())
}

// админ к себе
func TestUserInfo_AdminSelf(t *testing.T) {
	ctx, st := suite.New(t)

	respLogin, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		AppId:    ssoId,
		Login:    adminLogin,
		Password: adminPass,
	})
	require.NoError(t, err)

	token := respLogin.GetToken()
	require.NotEmpty(t, token)

	// Парсим токен, чтобы получить uid админа
	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(ssoSecret), nil
	})
	require.NoError(t, err)

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	require.True(t, ok)

	adminUserIdStr := claims["uid"].(string)
	adminUserId, err := strconv.ParseInt(adminUserIdStr, 10, 64)
	require.NoError(t, err)

	// Создаем контекст с токеном для авторизованного запроса
	ctxWithToken := metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "Bearer "+token))

	respUserInfo, err := st.AuthClient.UserInfo(ctxWithToken, &ssov1.UserInfoRequest{
		UserId: adminUserId,
	})
	require.NoError(t, err)
	assert.Equal(t, adminUserId, respUserInfo.GetUserId())
	assert.Equal(t, adminLogin, respUserInfo.GetLogin())
	assert.True(t, respUserInfo.GetIsAdmin())
}

// TODO: админ к пользователю
func TestUserInfo_AdminToUser(t *testing.T) {
	ctx, st := suite.New(t)

	// Регистрируем обычного пользователя
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

	// Логинимся как админ
	respLogin, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		AppId:    ssoId,
		Login:    adminLogin,
		Password: adminPass,
	})
	require.NoError(t, err)

	token := respLogin.GetToken()
	require.NotEmpty(t, token)

	// Создаем контекст с токеном для авторизованного запроса
	ctxWithToken := metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "Bearer "+token))

	// Админ запрашивает информацию о пользователе
	respUserInfo, err := st.AuthClient.UserInfo(ctxWithToken, &ssov1.UserInfoRequest{
		UserId: respRegister.GetUserId(),
	})
	require.NoError(t, err)
	assert.Equal(t, respRegister.GetUserId(), respUserInfo.GetUserId())
	assert.Equal(t, login, respUserInfo.GetLogin())
	assert.Equal(t, email, respUserInfo.GetEmail())
	assert.Equal(t, fullName, respUserInfo.GetFullName())
	assert.False(t, respUserInfo.GetIsAdmin())
}
