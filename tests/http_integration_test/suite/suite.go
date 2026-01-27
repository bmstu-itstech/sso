package suite

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/bmstu-itstech/sso/internal/domain/models"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
)

const (
	defaultBaseURL = "http://localhost:8080"
	httpTimeout    = 5 * time.Second
	waitTimeout    = 30 * time.Second
)

type Suite struct {
	*testing.T
	Ctx     context.Context
	Client  *http.Client
	BaseURL string
}

func New(t *testing.T) *Suite {
	t.Helper()
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), waitTimeout)
	t.Cleanup(func() {
		t.Helper()
		cancel()
	})

	baseURL := os.Getenv("SSO_HTTP_BASE_URL")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	st := &Suite{
		T:       t,
		Ctx:     ctx,
		Client:  &http.Client{Timeout: httpTimeout},
		BaseURL: baseURL,
	}

	st.waitForPing()
	return st
}

func (s *Suite) waitForPing() {
	s.Helper()

	deadline := time.Now().Add(waitTimeout)
	for time.Now().Before(deadline) {
		req, _ := http.NewRequestWithContext(s.Ctx, http.MethodGet, s.BaseURL+"/api/v1/ping", nil)
		resp, err := s.Client.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(500 * time.Millisecond)
	}

	s.T.Fatalf("sso app is not ready: ping %s/api/v1/ping did not return 200 within %s", s.BaseURL, waitTimeout)
}

// DoJSON - низкоуровневый вызов JSON endpoint (нужен для негативных тестов).
func (s *Suite) DoJSON(method, path string, body any, headers map[string]string) (*http.Response, []byte) {
	s.Helper()

	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(s.T, err)
		r = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(s.Ctx, method, s.BaseURL+path, r)
	require.NoError(s.T, err)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := s.Client.Do(req)
	require.NoError(s.T, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(s.T, err)
	require.NoError(s.T, resp.Body.Close())

	return resp, respBody
}

type TestUser struct {
	Login    string
	Password string
	Email    string
	FullName string
	UserID   int64
	Token    string
}

func (s *Suite) NewRandomUser() TestUser {
	s.Helper()
	return TestUser{
		Email:    gofakeit.Email(),
		Login:    fmt.Sprintf("%s_%d", gofakeit.Username(), gofakeit.Int64()),
		FullName: gofakeit.Name(),
		Password: gofakeit.Password(true, true, true, true, false, 12),
	}
}

func (s *Suite) RegisterUser(u TestUser) TestUser {
	s.Helper()

	resp, body := s.DoJSON(http.MethodPost, "/api/v1/register", models.RegisterRequest{
		Login:    u.Login,
		Password: u.Password,
		Email:    u.Email,
		FullName: u.FullName,
	}, nil)
	require.Equal(s.T, http.StatusOK, resp.StatusCode, "register response: %s", string(body))

	var out models.RegisterResponse
	require.NoError(s.T, json.Unmarshal(body, &out))
	require.NotZero(s.T, out.UserId)

	u.UserID = out.UserId
	return u
}

func (s *Suite) LoginUser(u TestUser, appID int32) TestUser {
	s.Helper()

	resp, body := s.DoJSON(http.MethodPost, "/api/v1/login", models.LoginRequest{
		AppId:    appID,
		Login:    u.Login,
		Password: u.Password,
	}, nil)
	require.Equal(s.T, http.StatusOK, resp.StatusCode, "login response: %s", string(body))

	var out models.LoginResponse
	require.NoError(s.T, json.Unmarshal(body, &out))
	require.NotEmpty(s.T, out.Token)

	u.Token = out.Token
	return u
}
