package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strconv"
)

type JWTConfig struct {
	JWT_ALGORITHM         string
	JWT_PRIVATE_KEY       *rsa.PrivateKey
	AccessTokenMinutesTTL int32
	RefreshTokenDaysTTL   int64
	MaxUserSessions       int
}

type APIConfig struct {
	API_HOST               string
	API_PORT               int
	DASHBOARD_FRONTEND_URI string
}

type OauthProviderConfig struct {
	DISCORD_AUTH_CLIENT_ID     string
	DISCORD_AUTH_CLIENT_SECRET string
	DISCORD_AUTH_REDIRECT_URI  string
}

type RedisConfig struct {
	REDIS_PORT     int
	REDIS_HOST     string
	REDIS_PASSWORD string
	REDIS_DB int
	REDIS_MAX_RETRIES int
}

var (
	JWT JWTConfig
	API   APIConfig
	OAuth OauthProviderConfig
	Redis RedisConfig
)

func Init() error {
	JWT.JWT_ALGORITHM = getEnv("JWT_ALGORITHM", "RS256")
	
	b64Key := getEnv("JWT_PRIVATE_KEY", "")
	if b64Key == "" {
		return errors.New("environment variable JWT_PRIVATE_KEY is required")
	}

	privateKey, err := ParseRSAPrivateKeyFromBase64(b64Key)
	if err != nil {
		return fmt.Errorf("failed to parse JWT_PRIVATE_KEY: %w", err)
	}

	JWT.JWT_PRIVATE_KEY = privateKey
	JWT.AccessTokenMinutesTTL = getEnvInt32("ACCESS_TOKEN_MINUTES_TTL", 2)
	JWT.RefreshTokenDaysTTL = getEnvInt64("REFRESH_TOKEN_DAYS_TTL", 30)
	JWT.MaxUserSessions = getEnvInt("MAX_USER_SESSIONS", 10)

	API.API_HOST = getEnv("API_HOST", "0.0.0.0")
	API.API_PORT = getEnvInt("API_PORT", 8080)
	API.DASHBOARD_FRONTEND_URI = getEnv("DASHBOARD_FRONTEND_URI", "http://localhost:3000")

	OAuth.DISCORD_AUTH_CLIENT_ID = getEnv("DISCORD_AUTH_CLIENT_ID", "")
	if OAuth.DISCORD_AUTH_CLIENT_ID == "" {
		return errors.New("environment variable DISCORD_AUTH_CLIENT_ID is required")
	}
	OAuth.DISCORD_AUTH_CLIENT_SECRET = getEnv("DISCORD_AUTH_CLIENT_SECRET", "")
	if OAuth.DISCORD_AUTH_CLIENT_SECRET == "" {
		return errors.New("environment variable DISCORD_AUTH_CLIENT_SECRET is required")
	}
	OAuth.DISCORD_AUTH_REDIRECT_URI = getEnv("DISCORD_AUTH_REDIRECT_URI", "http://localhost:8080/auth/discord/callback")

	Redis.REDIS_HOST = getEnv("REDIS_HOST", "127.0.0.1")
	Redis.REDIS_PORT = getEnvInt("REDIS_PORT", 6379)
	Redis.REDIS_PASSWORD = getEnv("REDIS_PASSWORD", "")
	Redis.REDIS_DB = getEnvInt("REDIS_DB", 0)
	Redis.REDIS_MAX_RETRIES = getEnvInt("REDIS_MAX_RETRIES", 5)

	return nil
}

func ParseRSAPrivateKeyFromBase64(b64Str string) (*rsa.PrivateKey, error) {
	if b64Str == "" {
		return nil, errors.New("base64 string is empty")
	}

	decoded, err := base64.StdEncoding.DecodeString(b64Str)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(b64Str)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64: %w", err)
		}
	}

	if block, _ := pem.Decode(decoded); block != nil {
		return parseRSAFromDER(block.Bytes)
	}

	return parseRSAFromDER(decoded)
}

func parseRSAFromDER(der []byte) (*rsa.PrivateKey, error) {
	if key, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		return key, nil
	}

	if keyIface, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		if key, ok := keyIface.(*rsa.PrivateKey); ok {
			return key, nil
		}
		return nil, errors.New("PKCS8 key is not RSA")
	}

	return nil, errors.New("failed to parse DER: unsupported key format (expected PKCS1 or PKCS8 RSA)")
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	valStr := getEnv(key, "")
	if valStr == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultValue
	}
	return val
}

func getEnvInt32(key string, defaultValue int32) int32 {
	valStr := getEnv(key, "")
	if valStr == "" {
		return defaultValue
	}
	val, err := strconv.ParseInt(valStr, 10, 32)
	if err != nil {
		return defaultValue
	}
	return int32(val)
}

func getEnvInt64(key string, defaultValue int64) int64 {
	valStr := getEnv(key, "")
	if valStr == "" {
		return defaultValue
	}
	val, err := strconv.ParseInt(valStr, 10, 64)
	if err != nil {
		return defaultValue
	}
	return val
}