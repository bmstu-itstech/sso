package suite

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bmstu-itstech/sso/internal/config"
	"github.com/bmstu-itstech/sso/internal/domain/models"
	http_server "github.com/bmstu-itstech/sso/internal/http"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

const (
	timeout = 5 * time.Second
)

type Suite struct {
	*testing.T
	Ctx    context.Context
	Router *gin.Engine
	Auth   *AuthMock
}

func New(t *testing.T) *Suite {
	t.Helper()
	t.Parallel()

	gin.SetMode(gin.TestMode)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(func() {
		t.Helper()
		cancel()
	})

	cfg := &config.Config{}
	mock := &AuthMock{}
	srv := http_server.New(mock, cfg)

	return &Suite{
		T:      t,
		Ctx:    ctx,
		Router: srv.InitRouter(),
		Auth:   mock,
	}
}

// DoJSON выполняет HTTP-запрос с JSON-телом (если body != nil) и возвращает recorder.
func (s *Suite) DoJSON(method, path string, body any, headers map[string]string) *httptest.ResponseRecorder {
	s.Helper()

	var reqBody *bytes.Reader
	if body == nil {
		reqBody = bytes.NewReader(nil)
	} else {
		b, err := json.Marshal(body)
		require.NoError(s.T, err)
		reqBody = bytes.NewReader(b)
	}

	req := httptest.NewRequest(method, path, reqBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	rr := httptest.NewRecorder()
	s.Router.ServeHTTP(rr, req)
	return rr
}

func (s *Suite) DecodeJSON(rr *httptest.ResponseRecorder, v any) {
	s.Helper()
	require.NotEmpty(s.T, rr.Body.Bytes())
	require.NoError(s.T, json.Unmarshal(rr.Body.Bytes(), v))
}

// ----- Mock ----

type AuthMock struct {
	LoginFn          func(ctx context.Context, appId int32, login string, password string) (string, error)
	RegisterFn       func(ctx context.Context, login, password, email, fullName string) (int64, error)
	UpdatePasswordFn func(ctx context.Context, id int64, newPassword string) error
	DeleteUserFn     func(ctx context.Context, userId int64) error
	SignInFn         func(ctx context.Context, token string, appId int32) (models.TokenInfo, error)
	IsAdminFn        func(ctx context.Context, userId int64) (bool, error)
	UserInfoFn       func(ctx context.Context, id int64) (models.UserServices, error)
	UsersAllFn       func(ctx context.Context) ([]models.UserServices, error)
	UpdateTokenAppFn func(ctx context.Context, token string, appId int32) (string, error)

	LoginCalls          int
	RegisterCalls       int
	SignInCalls         int
	IsAdminCalls        int
	UpdateTokenAppCalls int
}

func (m *AuthMock) Login(ctx context.Context, appId int32, login string, password string) (string, error) {
	m.LoginCalls++
	if m.LoginFn == nil {
		return "", errors.New("LoginFn is not set")
	}
	return m.LoginFn(ctx, appId, login, password)
}

func (m *AuthMock) RegisterNewUser(ctx context.Context, login, password, email, fullName string) (int64, error) {
	m.RegisterCalls++
	if m.RegisterFn == nil {
		return 0, errors.New("RegisterFn is not set")
	}
	return m.RegisterFn(ctx, login, password, email, fullName)
}

func (m *AuthMock) UpdatePassword(ctx context.Context, id int64, newPassword string) error {
	if m.UpdatePasswordFn == nil {
		return errors.New("UpdatePasswordFn is not set")
	}
	return m.UpdatePasswordFn(ctx, id, newPassword)
}

func (m *AuthMock) DeleteUser(ctx context.Context, userId int64) error {
	if m.DeleteUserFn == nil {
		return errors.New("DeleteUserFn is not set")
	}
	return m.DeleteUserFn(ctx, userId)
}

func (m *AuthMock) SignIn(ctx context.Context, token string, appId int32) (models.TokenInfo, error) {
	m.SignInCalls++
	if m.SignInFn == nil {
		return models.TokenInfo{}, errors.New("SignInFn is not set")
	}
	return m.SignInFn(ctx, token, appId)
}

func (m *AuthMock) IsAdmin(ctx context.Context, userId int64) (bool, error) {
	m.IsAdminCalls++
	if m.IsAdminFn == nil {
		return false, errors.New("IsAdminFn is not set")
	}
	return m.IsAdminFn(ctx, userId)
}

func (m *AuthMock) UserInfo(ctx context.Context, id int64) (models.UserServices, error) {
	if m.UserInfoFn == nil {
		return models.UserServices{}, errors.New("UserInfoFn is not set")
	}
	return m.UserInfoFn(ctx, id)
}

func (m *AuthMock) UsersAll(ctx context.Context) ([]models.UserServices, error) {
	if m.UsersAllFn == nil {
		return nil, errors.New("UsersAllFn is not set")
	}
	return m.UsersAllFn(ctx)
}

func (m *AuthMock) UpdateTokenApp(ctx context.Context, token string, appId int32) (string, error) {
	m.UpdateTokenAppCalls++
	if m.UpdateTokenAppFn == nil {
		return "", errors.New("UpdateTokenAppFn is not set")
	}
	return m.UpdateTokenAppFn(ctx, token, appId)
}

// ensure interface compliance at compile-time (no globals, so in a function)
func (s *Suite) RequireAuthInterface() {
	s.Helper()
	var _ http_server.Auth = (*AuthMock)(nil)
	var _ = http.MethodGet
}
