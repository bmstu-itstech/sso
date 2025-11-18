package snake_case

import (
	"google.golang.org/grpc/metadata"
	"strconv"
	"testing"

	ssov1 "github.com/BOBAvov/protos_sso/gen/go/sso"
	"github.com/bmstu-itstech/sso/snake_case/suite"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// тест обычное использование
func TestIsAdminIsNot(t *testing.T) {
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

	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(ssoSecret), nil
	})
	require.NoError(t, err)

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	require.True(t, ok)

	t.Log(respRegister.GetUserId(), claims["uid"])
	assert.Equal(t, strconv.FormatInt(respRegister.GetUserId(), 10), claims["uid"].(string))

	// Создаем контекст с токеном для авторизованного запроса
	ctxWithToken := metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "Bearer "+respLogin.GetToken()))
	respIsAdmin, err := st.AuthClient.IsAdmin(ctxWithToken, &ssov1.IsAdminRequest{
		UserId: respRegister.GetUserId(),
	})
	require.NoError(t, err)
	assert.False(t, respIsAdmin.IsAdmin)
}
