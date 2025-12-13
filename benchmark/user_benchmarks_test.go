package benchmark

import (
	"context"
	"strconv"
	"testing"

	ssov1 "github.com/BOBAvov/protos_sso/gen/go/sso"
	"github.com/bmstu-itstech/sso/benchmark/suite"
	"github.com/brianvoe/gofakeit/v6"
	"google.golang.org/grpc/metadata"
)

type UserBenchmarks struct {
	Login    string
	Password string
	UID      int64
	Token    string
}

// prepareUser создаёт пользователя и логинится один раз — возвращает подготовленные данные
func prepareUser(ctxCtx context.Context, st *suite.Suite) (UserBenchmarks, error) {
	// Подготовка пользователя
	email := gofakeit.Email()
	login := gofakeit.Username()
	fullName := gofakeit.Name()
	password := gofakeit.Password(true, true, true, true, false, passDefaultLen)

	respReg, err := st.AuthClient.Register(ctxCtx, &ssov1.RegisterRequest{
		Login:    login,
		Password: password,
		Email:    email,
		FullName: fullName,
	})
	if err != nil {
		return UserBenchmarks{}, err
	}

	// логин (чтобы сервис мог выдать токен)
	respLogin, err := st.AuthClient.Login(ctxCtx, &ssov1.LoginRequest{
		AppId:    int32(appId),
		Login:    login,
		Password: password,
	})
	if err != nil {
		return UserBenchmarks{}, err
	}

	var token string
	if respLogin != nil {
		token = respLogin.GetToken()
	}

	return UserBenchmarks{
		Login:    login,
		Password: password,
		UID:      respReg.GetUserId(),
		Token:    token,
	}, nil
}

// Бенчмарк регистрации
func BenchmarkRegister(b *testing.B) {
	ctx, st := suite.New(b)
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		email := gofakeit.Email()
		login := gofakeit.Username() + strconv.Itoa(i)
		fullName := gofakeit.Name()
		password := gofakeit.Password(true, true, true, true, false, passDefaultLen)

		_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
			Login:    login,
			Password: password,
			Email:    email,
			FullName: fullName,
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Бенчмарк: регистрация + логин
func BenchmarkRegisterLogin(b *testing.B) {
	ctx, st := suite.New(b)
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		email := gofakeit.Email()
		login := gofakeit.Username() + strconv.Itoa(i)
		fullName := gofakeit.Name()
		password := gofakeit.Password(true, true, true, true, false, passDefaultLen)

		_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
			Login:    login,
			Password: password,
			Email:    email,
			FullName: fullName,
		})
		if err != nil {
			b.Fatal(err)
		}

		_, err = st.AuthClient.Login(ctx, &ssov1.LoginRequest{
			AppId:    int32(appId),
			Login:    login,
			Password: password,
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Бенчмарк логина для заранее зарегистрированного пользователя
func BenchmarkLogin(b *testing.B) {
	ctx, st := suite.New(b)
	b.ReportAllocs()

	// Подготовка: один пользователь
	email := gofakeit.Email()
	login := gofakeit.Username()
	fullName := gofakeit.Name()
	password := gofakeit.Password(true, true, true, true, false, passDefaultLen)

	_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Login:    login,
		Password: password,
		Email:    email,
		FullName: fullName,
	})
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
			AppId:    int32(appId),
			Login:    login,
			Password: password,
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Бенчмарк обновления JWT токена пользователем
// Использует заранее подготовленного пользователя (prepareUser)
func BenchmarkUpdateToken_User(b *testing.B) {
	ctx, st := suite.New(b)
	b.ReportAllocs()

	ub, err := prepareUser(ctx, st)
	if err != nil {
		b.Fatal(err)
	}

	// используем заранее полученный токен (и UID) — не регистрируем/логинимся в цикле
	ctxWithToken := metadata.NewOutgoingContext(ctx, metadata.Pairs(
		"authorization", "Bearer "+ub.Token,
		"x-user-id", strconv.FormatInt(ub.UID, 10),
	))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := st.AuthClient.UpdateToken(ctxWithToken, &ssov1.UpdateTokenRequest{AppId: int32(appId)})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Бенчмарк обновления пароля пользователем (self)
// Использует заранее подготовленного пользователя (prepareUser)
func BenchmarkUpdatePassword_UserSelf(b *testing.B) {
	ctx, st := suite.New(b)
	b.ReportAllocs()

	ub, err := prepareUser(ctx, st)
	if err != nil {
		b.Fatal(err)
	}

	// используем заранее полученный токен (и UID)
	ctxWithToken := metadata.NewOutgoingContext(ctx, metadata.Pairs(
		"authorization", "Bearer "+ub.Token,
		"x-user-id", strconv.FormatInt(ub.UID, 10),
	))

	newPassword := gofakeit.Password(true, true, true, true, false, passDefaultLen)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := st.AuthClient.UpdatePassword(ctxWithToken, &ssov1.UpdatePasswordRequest{
			UserId:      ub.UID,
			NewPassword: newPassword,
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Бенчмарк получения информации о пользователе (self)
// Использует заранее подготовленного пользователя (prepareUser)
func BenchmarkUserInfo_UserSelf(b *testing.B) {
	ctx, st := suite.New(b)
	b.ReportAllocs()

	ub, err := prepareUser(ctx, st)
	if err != nil {
		b.Fatal(err)
	}

	// используем заранее полученный токен (и UID)
	ctxWithToken := metadata.NewOutgoingContext(ctx, metadata.Pairs(
		"authorization", "Bearer "+ub.Token,
		"x-user-id", strconv.FormatInt(ub.UID, 10),
	))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := st.AuthClient.UserInfo(ctxWithToken, &ssov1.UserInfoRequest{UserId: ub.UID})
		if err != nil {
			b.Fatal(err)
		}
	}
}
