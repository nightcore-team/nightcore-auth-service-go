package discord

import "nightcore-team/nightcore-auth-service-go/internal/domain/entity"

type OauthProvider struct{}

func NewOauthProvider() *OauthProvider {
	return &OauthProvider{}
}

func (*OauthProvider) ExchangeCode(code string) *entity.DiscordTokenData {

}

func (*OauthProvider) GetUserInfo(accessToken string) *entity.DiscordUserData {

}