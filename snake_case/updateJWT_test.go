package snake_case

import (
	ssov1 "github.com/BOBAvov/protos_sso/gen/go/sso"
	"github.com/bmstu-itstech/sso/snake_case/suite"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
	"strconv"
	"testing"
)

// обнавление токена пользователя
func TestUpdateToken_User(t *testing.T) {
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
		AppId:    appId,
		Login:    login,
		Password: password,
	})
	require.NoError(t, err)

	oldToken := respLogin.GetToken()
	require.NotEmpty(t, oldToken)

	// Парсим старый токен для проверки
	oldTokenParsed, err := jwt.Parse(oldToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})
	require.NoError(t, err)

	oldClaims, ok := oldTokenParsed.Claims.(jwt.MapClaims)
	require.True(t, ok)

	// Создаем контекст с токеном для авторизованного запроса
	ctxWithToken := metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "Bearer "+oldToken))

	// Обновляем токен
	respUpdateToken, err := st.AuthClient.UpdateToken(ctxWithToken, &emptypb.Empty{})
	require.NoError(t, err)
	assert.NotEmpty(t, respUpdateToken.GetToken())

	newToken := respUpdateToken.GetToken()
	require.NotEmpty(t, newToken)

	// Парсим новый токен и проверяем данные
	newTokenParsed, err := jwt.Parse(newToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})
	require.NoError(t, err)

	newClaims, ok := newTokenParsed.Claims.(jwt.MapClaims)
	require.True(t, ok)

	// Проверяем, что данные пользователя совпадают
	assert.Equal(t, strconv.FormatInt(respRegister.GetUserId(), 10), newClaims["uid"].(string))
	assert.Equal(t, oldClaims["uid"], newClaims["uid"])
	assert.Equal(t, appId, int(newClaims["app_id"].(float64)))
}
