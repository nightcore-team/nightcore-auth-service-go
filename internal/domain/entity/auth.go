package entity

type TokenExchangeResult struct {
	AccessToken  string
	RefreshToken string
}

type Session struct {
	UserID    int64
	IpAddress string
}

type DiscordTokenData struct {
	Access_token  string
	Token_type    string
	Expires_in    int64
	Refresh_token string
	Scope         string
}

type DiscordUserData struct {
	ID int64
}