package grpc_test

import (
	"github.com/bmstu-itstech/sso/tests/grpc_test/suite"
	"math/rand"
	"strconv"
	"testing"

	ssov1 "github.com/BOBAvov/protos_sso/gen/go/sso"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

// админ удаляет пользователя
func TestDeleteUser_AdminDeletesUser(t *testing.T) {
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

	// Админ удаляет пользователя
	respDeleteUser, err := st.AuthClient.RemoveUser(ctxWithToken, &ssov1.RemoveUserRequest{
		UserId: respRegister.GetUserId(),
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respDeleteUser.GetMessage())

	// Проверяем, что пользователь больше не может залогиниться
	_, err = st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		AppId:    appId,
		Login:    login,
		Password: password,
	})
	require.Error(t, err)
}

// пользователь удаляет себя
func TestDeleteUser_UserDeletesSelf(t *testing.T) {
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

	// Пользователь удаляет себя
	respDeleteUser, err := st.AuthClient.RemoveUser(ctxWithToken, &ssov1.RemoveUserRequest{
		UserId: respRegister.GetUserId(),
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respDeleteUser.GetMessage())

	// Проверяем, что пользователь больше не может залогиниться
	_, err = st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		AppId:    appId,
		Login:    login,
		Password: password,
	})
	require.Error(t, err)
}

// админ удаляет не существующего пользователя
func TestDeleteUser_AdminDeletesNonExistentUser(t *testing.T) {
	ctx, st := suite.New(t)

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

	// Админ пытается удалить несуществующего пользователя
	fakeUserId := rand.Int63()
	_, err = st.AuthClient.RemoveUser(ctxWithToken, &ssov1.RemoveUserRequest{
		UserId: fakeUserId,
	})
	require.NoError(t, err)
}
