package discord

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"nightcore-team/nightcore-auth-service-go/config"
	"nightcore-team/nightcore-auth-service-go/internal/domain"
	"nightcore-team/nightcore-auth-service-go/internal/domain/entity"
	"strings"
	"time"
)

type OauthProvider struct{
	httpClient *http.Client
}

func NewOauthProvider() *OauthProvider {
	return &OauthProvider{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (p *OauthProvider) ExchangeCode(ctx context.Context, code string) (*entity.DiscordTokenData, *domain.AppError) {

	data := url.Values{
		"client_id":     {config.OAuth.DISCORD_AUTH_CLIENT_ID},
		"client_secret": {config.OAuth.DISCORD_AUTH_CLIENT_SECRET},
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {config.OAuth.DISCORD_AUTH_REDIRECT_URI},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://discord.com/api/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, domain.ErrOauthExchangeFailed.WithCause(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, domain.ErrOauthExchangeFailed.WithCause(err)
	}
	defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
	body, _ := io.ReadAll(resp.Body)
	return nil, domain.ErrOauthExchangeFailed.WithCause(fmt.Errorf("discord returned %d: %s", resp.StatusCode, body))
}

	var tokenData entity.DiscordTokenData
	if err := json.NewDecoder(resp.Body).Decode(&tokenData); err != nil {
		return nil, domain.ErrOauthExchangeFailed.WithCause(err)
	}

	return &tokenData, nil
}

func (p *OauthProvider) GetUserInfo(ctx context.Context, accessToken string) (*entity.DiscordUserData, *domain.AppError) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://discord.com/api/users/@me", nil)
	if err != nil {
		return nil, domain.ErrOauthExchangeFailed.WithCause(err)
	}
	req.Header.Set("Authorization", "Bearer " + accessToken)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, domain.ErrOauthExchangeFailed.WithCause(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, domain.ErrOauthExchangeFailed.WithCause(fmt.Errorf("discord returned %d: %s", resp.StatusCode, body))
	}

	var userData entity.DiscordUserData
	if err := json.NewDecoder(resp.Body).Decode(&userData); err != nil {
		return nil, domain.ErrOauthExchangeFailed.WithCause(err)
	}

	return &userData, nil
}