package rds

import (
	"fmt"
	"nightcore-team/nightcore-auth-service-go/config"

	"github.com/redis/go-redis/v9"
)

func  NewRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", config.Redis.REDIS_HOST, config.Redis.REDIS_PORT),
		Password: config.Redis.REDIS_PASSWORD,
		DB: config.Redis.REDIS_DB,
		MaxRetries: config.Redis.REDIS_MAX_RETRIES,
	})
}