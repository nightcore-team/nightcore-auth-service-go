package discord

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"nightcore-team/nightcore-auth-service-go/config"

	"github.com/stretchr/testify/require"
)

type mockTransport struct {
	statusCode int
	body       string
	err        error
}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.err != nil {
		return nil, t.err
	}
	return &http.Response{
		StatusCode: t.statusCode,
		Body:       io.NopCloser(strings.NewReader(t.body)),
		Header:     make(http.Header),
	}, nil
}

func setupConfig() {
	config.OAuth.DISCORD_AUTH_CLIENT_ID = "test-client-id"
	config.OAuth.DISCORD_AUTH_CLIENT_SECRET = "test-client-secret"
	config.OAuth.DISCORD_AUTH_REDIRECT_URI = "http://localhost/callback"
}

func newTestProvider(statusCode int, body string, err error) *OauthProvider {
	return &OauthProvider{
		httpClient: &http.Client{
			Transport: &mockTransport{
				statusCode: statusCode,
				body:       body,
				err:        err,
			},
		},
	}
}

func TestOauthProvider_ExchangeCode(t *testing.T) {
	setupConfig()

	t.Run("success", func(t *testing.T) {
		provider := newTestProvider(http.StatusOK, `{"access_token": "acc", "refresh_token": "ref"}`, nil)

		res, appErr := provider.ExchangeCode(context.Background(), "test-code")

		require.Nil(t, appErr)
		require.NotNil(t, res)
	})

	t.Run("http error status", func(t *testing.T) {
		provider := newTestProvider(http.StatusBadRequest, `{"error": "invalid_grant"}`, nil)

		res, appErr := provider.ExchangeCode(context.Background(), "bad-code")

		require.Nil(t, res)
		require.NotNil(t, appErr)
	})

	t.Run("invalid json", func(t *testing.T) {
		provider := newTestProvider(http.StatusOK, `invalid json`, nil)

		res, appErr := provider.ExchangeCode(context.Background(), "test-code")

		require.Nil(t, res)
		require.NotNil(t, appErr)
	})

	t.Run("network error", func(t *testing.T) {
		provider := newTestProvider(0, "", errors.New("network error"))

		res, appErr := provider.ExchangeCode(context.Background(), "test-code")

		require.Nil(t, res)
		require.NotNil(t, appErr)
	})

	t.Run("context canceled", func(t *testing.T) {
		provider := newTestProvider(0, "", nil)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		res, appErr := provider.ExchangeCode(ctx, "test-code")

		require.Nil(t, res)
		require.NotNil(t, appErr)
	})
}

func TestOauthProvider_GetUserInfo(t *testing.T) {
	setupConfig()

	t.Run("success", func(t *testing.T) {
		provider := newTestProvider(http.StatusOK, `{"id": "123456789", "username": "testuser"}`, nil)

		res, appErr := provider.GetUserInfo(context.Background(), "valid-token")

		require.Nil(t, appErr)
		require.NotNil(t, res)
	})

	t.Run("http error status", func(t *testing.T) {
		provider := newTestProvider(http.StatusUnauthorized, `{"message": "401: Unauthorized"}`, nil)

		res, appErr := provider.GetUserInfo(context.Background(), "bad-token")

		require.Nil(t, res)
		require.NotNil(t, appErr)
	})

	t.Run("invalid json", func(t *testing.T) {
		provider := newTestProvider(http.StatusOK, `not json`, nil)

		res, appErr := provider.GetUserInfo(context.Background(), "valid-token")

		require.Nil(t, res)
		require.NotNil(t, appErr)
	})

	t.Run("network error", func(t *testing.T) {
		provider := newTestProvider(0, "", errors.New("connection refused"))

		res, appErr := provider.GetUserInfo(context.Background(), "valid-token")

		require.Nil(t, res)
		require.NotNil(t, appErr)
	})

	t.Run("context canceled", func(t *testing.T) {
		provider := newTestProvider(0, "", nil)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		res, appErr := provider.GetUserInfo(ctx, "valid-token")

		require.Nil(t, res)
		require.NotNil(t, appErr)
	})
}