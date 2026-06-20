package rds

import (
	"fmt"
	"nightcore-team/nightcore-auth-service-go/config"

	"github.com/go-redis/redis"
)

func  NewRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", config.Redis.REDIS_HOST, config.Redis.REDIS_PORT),
		Password: config.Redis.REDIS_PASSWORD,
		DB: 0,
		MaxRetries: 5,
	})
}