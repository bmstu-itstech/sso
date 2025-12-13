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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// пользователь пытается получить доступ
func TestUsersInfo_UserTriesToGetAccess(t *testing.T) {
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

	// Пользователь пытается получить список всех пользователей
	_, err = st.AuthClient.UsersInfo(ctxWithToken, &emptypb.Empty{})
	require.Error(t, err)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

// админ получает список пользователь
func TestUsersInfo_AdminGetsUsersList(t *testing.T) {
	ctx, st := suite.New(t)

	// Регистрируем нескольких пользователей
	users := make([]*ssov1.RegisterRequest, 3)
	for i := 0; i < 3; i++ {
		users[i] = &ssov1.RegisterRequest{
			Login:    gofakeit.Username(),
			Password: randomFakePassword(),
			Email:    gofakeit.Email(),
			FullName: gofakeit.Name(),
		}
		_, err := st.AuthClient.Register(ctx, users[i])
		require.NoError(t, err)
	}

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

	// Админ получает список всех пользователей
	respUsersInfo, err := st.AuthClient.UsersInfo(ctxWithToken, &emptypb.Empty{})
	require.NoError(t, err)
	assert.NotNil(t, respUsersInfo)
	assert.NotEmpty(t, respUsersInfo.GetUsers())

	// Проверяем, что в списке есть админ
	foundAdmin := false
	for _, user := range respUsersInfo.GetUsers() {
		if user.GetLogin() == adminLogin && user.GetIsAdmin() {
			foundAdmin = true
			break
		}
	}
	assert.True(t, foundAdmin, "Admin should be in the users list")
}
