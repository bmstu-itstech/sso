package http_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/bmstu-itstech/sso/internal/domain/models"
	"github.com/bmstu-itstech/sso/tests/http_test/suite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	appID int32 = 0
)

func TestLogin_HappyPath(t *testing.T) {
	st := suite.New(t)

	st.Auth.LoginFn = func(_ctx context.Context, appId int32, login, password string) (string, error) {
		assert.Equal(t, appID, appId)
		assert.Equal(t, "alice", login)
		assert.Equal(t, "pass", password)
		return "jwt-token", nil
	}

	rr := st.DoJSON(http.MethodPost, "/api/v1/login", models.LoginRequest{
		AppId:    appID,
		Login:    "alice",
		Password: "pass",
	}, nil)
	require.Equal(t, http.StatusOK, rr.Code)

	var resp models.LoginResponse

	st.DecodeJSON(rr, &resp)
	t.Logf("resp=%s", resp.Token)
	require.Equal(t, "jwt-token", resp.Token)
	require.Equal(t, 1, st.Auth.LoginCalls)
}

func TestLogin_InvalidCredentials(t *testing.T) {
	st := suite.New(t)

	st.Auth.LoginFn = func(_ctx context.Context, _ int32, _, _ string) (string, error) {
		return "", errors.New("invalid credentials")
	}

	rr := st.DoJSON(http.MethodPost, "/api/v1/login", models.LoginRequest{
		AppId:    appID,
		Login:    "alice",
		Password: "wrong",
	}, nil)
	require.Equal(t, http.StatusUnauthorized, rr.Code)
	require.Equal(t, 1, st.Auth.LoginCalls)
}

func TestRegister_HappyPath(t *testing.T) {
	st := suite.New(t)

	st.Auth.RegisterFn = func(_ctx context.Context, login, password, email, fullName string) (int64, error) {
		assert.Equal(t, "bob", login)
		assert.Equal(t, "password123", password)
		assert.Equal(t, "bob@example.com", email)
		assert.Equal(t, "Bob B.", fullName)
		return 42, nil
	}

	rr := st.DoJSON(http.MethodPost, "/api/v1/register", models.RegisterRequest{
		Login:    "bob",
		Password: "password123",
		Email:    "bob@example.com",
		FullName: "Bob B.",
	}, nil)
	require.Equal(t, http.StatusOK, rr.Code)

	var resp models.RegisterResponse
	st.DecodeJSON(rr, &resp)
	require.Equal(t, int64(42), resp.UserId)
	require.Equal(t, 1, st.Auth.RegisterCalls)
}

func TestRegister_ValidationError(t *testing.T) {
	st := suite.New(t)

	// Здесь mock не должен вызываться: ошибка должна быть на уровне валидации input.
	st.Auth.RegisterFn = func(_ctx context.Context, _, _, _, _ string) (int64, error) {
		t.Fatalf("RegisterFn must not be called on validation error")
		return 0, nil
	}

	rr := st.DoJSON(http.MethodPost, "/api/v1/register", models.RegisterRequest{
		Login:    "bob",
		Password: "short", // min=8
		Email:    "not-an-email",
		FullName: "",
	}, nil)
	require.Equal(t, http.StatusBadRequest, rr.Code)
	require.Equal(t, 0, st.Auth.RegisterCalls)
}
