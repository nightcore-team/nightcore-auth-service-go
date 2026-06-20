package config

import (
	"errors"
	"os"
	"strconv"
)

type JWTConfig struct {
	JWT_ALGORITHM         string
	JWT_PRIVATE_KEY       string
	AccessTokenMinutesTTL int32
	RefreshTokenDaysTTL   int64
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
}

var (
	JWT JWTConfig
	API   APIConfig
	OAuth OauthProviderConfig
	Redis RedisConfig
)

func Init() error {
	JWT.JWT_ALGORITHM = getEnv("JWT_ALGORITHM", "RS256")
	JWT.JWT_PRIVATE_KEY = getEnv("JWT_PRIVATE_KEY", "")
	if JWT.JWT_PRIVATE_KEY == "" {
		return errors.New("environment variable JWT_PRIVATE_KEY is required")
	}
	JWT.AccessTokenMinutesTTL = getEnvInt32("ACCESS_TOKEN_MINUTES_TTL", 2)
	JWT.RefreshTokenDaysTTL = getEnvInt64("REFRESH_TOKEN_DAYS_TTL", 30)

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

	return nil
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