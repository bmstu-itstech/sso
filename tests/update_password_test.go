package tests

import (
	"strconv"
	"testing"

	ssov1 "github.com/BOBAvov/protos_sso/gen/go/sso"
	"github.com/bmstu-itstech/sso/tests/suite"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

// пользователь у себя
func TestUpdatePassword_UserSelf(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	login := gofakeit.Username()
	fullName := gofakeit.Name()
	password := randomFakePassword()
	newPassword := randomFakePassword()

	t.Log(email, login, fullName, password)

	respRegister, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Login:    login,
		Password: password,
		Email:    email,
		FullName: fullName,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respRegister.GetUserId())

	respLogin, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		AppId:    ssoId,
		Login:    login,
		Password: password,
	})
	require.NoError(t, err)

	token := respLogin.GetToken()
	require.NotEmpty(t, token)

	// Создаем контекст с токеном для авторизованного запроса
	ctxWithToken := metadata.NewOutgoingContext(ctx, metadata.Pairs("x-user-id", strconv.FormatInt(respRegister.GetUserId(), 10)))

	respUpdatePassword, err := st.AuthClient.UpdatePassword(ctxWithToken, &ssov1.UpdatePasswordRequest{
		UserId:      respRegister.GetUserId(),
		NewPassword: newPassword,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respUpdatePassword.GetMessage())

	// Проверяем, что теперь можно залогиниться с новым паролем
	respLoginNew, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		AppId:    ssoId,
		Login:    login,
		Password: newPassword,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respLoginNew.GetToken())
}

// админ у пользователя
func TestUpdatePassword_AdminToUser(t *testing.T) {
	ctx, st := suite.New(t)

	// Регистрируем обычного пользователя
	email := gofakeit.Email()
	login := gofakeit.Username()
	fullName := gofakeit.Name()
	password := randomFakePassword()
	newPassword := randomFakePassword()

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

	// Парсим токен, чтобы получить uid админа
	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(ssoSecret), nil
	})
	require.NoError(t, err)

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	require.True(t, ok)

	adminUserIdStr := claims["uid"].(string)

	// Создаем контекст с токеном для авторизованного запроса
	ctxWithToken := metadata.NewOutgoingContext(ctx, metadata.Pairs("x-user-id", adminUserIdStr))

	// Админ обновляет пароль пользователя
	respUpdatePassword, err := st.AuthClient.UpdatePassword(ctxWithToken, &ssov1.UpdatePasswordRequest{
		UserId:      respRegister.GetUserId(),
		NewPassword: newPassword,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respUpdatePassword.GetMessage())

	// Проверяем, что пользователь теперь может залогиниться с новым паролем
	respLoginNew, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		AppId:    ssoId,
		Login:    login,
		Password: newPassword,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respLoginNew.GetToken())
}
