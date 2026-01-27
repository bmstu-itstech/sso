package http_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/bmstu-itstech/sso/internal/domain/models"
	"github.com/bmstu-itstech/sso/tests/http_test/suite"
	"github.com/stretchr/testify/require"
)

const (
	bearerOK  = "Bearer ok"
	bearerBad = "Bearer bad"

	adminUID    = int64(1)
	nonAdminUID = int64(2)
)

func TestIsAdmin_HappyPath_AdminAsksSomeone(t *testing.T) {
	st := suite.New(t)

	st.Auth.SignInFn = func(_ context.Context, token string, appId int32) (models.TokenInfo, error) {
		require.Equal(t, "ok", token)
		return models.TokenInfo{Uid: adminUID, IsAdmin: true, AppId: appId}, nil
	}
	st.Auth.IsAdminFn = func(_ context.Context, userId int64) (bool, error) {
		require.Equal(t, int64(99), userId)
		return false, nil
	}

	rr := st.DoJSON(http.MethodGet, "/api/v1/user/is_admin/99", nil, map[string]string{
		"Authorization": bearerOK,
	})
	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, 1, st.Auth.SignInCalls)
}

func TestIsAdmin_Forbidden_NonAdminAsksOtherUser(t *testing.T) {
	st := suite.New(t)

	st.Auth.SignInFn = func(_ context.Context, token string, appId int32) (models.TokenInfo, error) {
		require.Equal(t, "ok", token)
		return models.TokenInfo{Uid: nonAdminUID, IsAdmin: false, AppId: appId}, nil
	}

	rr := st.DoJSON(http.MethodGet, "/api/v1/user/is_admin/99", nil, map[string]string{
		"Authorization": bearerOK,
	})
	require.Equal(t, http.StatusForbidden, rr.Code)
}

func TestUpdateToken_HappyPath(t *testing.T) {
	st := suite.New(t)

	st.Auth.SignInFn = func(_ context.Context, token string, appId int32) (models.TokenInfo, error) {
		require.Equal(t, "ok", token)
		return models.TokenInfo{Uid: adminUID, IsAdmin: true, AppId: appId}, nil
	}
	st.Auth.UpdateTokenAppFn = func(_ context.Context, oldToken string, appId int32) (string, error) {
		require.Equal(t, "ok", oldToken)
		require.Equal(t, int32(123), appId)
		return "new-token", nil
	}

	rr := st.DoJSON(http.MethodPost, "/api/v1/user/update_token", models.UpdateTokenRequest{AppId: 123}, map[string]string{
		"Authorization": bearerOK,
	})
	require.Equal(t, http.StatusOK, rr.Code)
}

func TestUpdateToken_ValidationError(t *testing.T) {
	st := suite.New(t)

	st.Auth.SignInFn = func(_ context.Context, token string, appId int32) (models.TokenInfo, error) {
		require.Equal(t, "ok", token)
		return models.TokenInfo{Uid: adminUID, IsAdmin: true, AppId: appId}, nil
	}
	st.Auth.UpdateTokenAppFn = func(_ context.Context, _ string, _ int32) (string, error) {
		t.Fatalf("UpdateTokenAppFn must not be called on validation error")
		return "", nil
	}

	rr := st.DoJSON(http.MethodPost, "/api/v1/user/update_token", models.UpdateTokenRequest{AppId: 0}, map[string]string{
		"Authorization": bearerOK,
	})
	require.Equal(t, http.StatusBadRequest, rr.Code)
	require.Equal(t, 0, st.Auth.UpdateTokenAppCalls)
}

func TestPrivateRoute_Unauthorized_NoHeader(t *testing.T) {
	st := suite.New(t)

	rr := st.DoJSON(http.MethodPost, "/api/v1/user/update_token", models.UpdateTokenRequest{AppId: 123}, nil)
	require.Equal(t, http.StatusUnauthorized, rr.Code)
	require.Equal(t, 0, st.Auth.SignInCalls)
}

func TestPrivateRoute_Unauthorized_BadToken(t *testing.T) {
	st := suite.New(t)

	st.Auth.SignInFn = func(_ context.Context, token string, appId int32) (models.TokenInfo, error) {
		require.Equal(t, "bad", token)
		return models.TokenInfo{}, errors.New("bad token")
	}

	rr := st.DoJSON(http.MethodPost, "/api/v1/user/update_token", models.UpdateTokenRequest{AppId: 123}, map[string]string{
		"Authorization": bearerBad,
	})
	require.Equal(t, http.StatusUnauthorized, rr.Code)
	require.Equal(t, 1, st.Auth.SignInCalls)
}
