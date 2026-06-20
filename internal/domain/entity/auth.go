package entity

import "encoding/json"

type TokenExchangeResult struct {
	AccessToken  string
	RefreshToken string
}

type Session struct {
	UserID    int64 `json:"user_id"`
	IpAddress string `json:"ip_address"`
}

func (s *Session) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}

func (s *Session) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
}

type DiscordTokenData struct {
	Access_token  string `json:"access_token"`
	Token_type    string `json:"token_type"`
	Expires_in    int64  `json:"expires_in"`
	Refresh_token string `json:"refresh_token"`
	Scope         string `json:"scope"`
}

type DiscordUserData struct {
	ID int64 `json:"id,string"`
}