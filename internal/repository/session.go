package repository

import (
	"context"
	"fmt"
	"nightcore-team/nightcore-auth-service-go/internal/domain"
	"nightcore-team/nightcore-auth-service-go/internal/domain/entity"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	GetDel(ctx context.Context, key string) *redis.StringCmd
	SMembers(ctx context.Context, key string) *redis.StringSliceCmd
	TxPipeline() redis.Pipeliner
}

type SessionRepository struct {
	client RedisClient
}

func NewSessionRepository(redisClient RedisClient) *SessionRepository {
	return &SessionRepository{
		client: redisClient,
	}
}

func (r *SessionRepository) sessionKey(refreshToken string) string {
	return fmt.Sprintf("session:%s", refreshToken)
}

func (r *SessionRepository) userSessionsKey(userID int64) string {
	return fmt.Sprintf("user_sessions:%d", userID)
}

func (r *SessionRepository) Get(ctx context.Context, refreshToken string) (*entity.Session, *domain.AppError) {
	cmd := r.client.Get(ctx, r.sessionKey(refreshToken))
	session := &entity.Session{}

	err := cmd.Scan(session)
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, domain.ErrUnknownRedis.WithCause(err)
	}

	return session, nil
}

func (r *SessionRepository) GetDel(ctx context.Context, refreshToken string) (*entity.Session, *domain.AppError) {
	cmd := r.client.GetDel(ctx, r.sessionKey(refreshToken))
	session := &entity.Session{}

	err := cmd.Scan(session)
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, domain.ErrUnknownRedis.WithCause(err)
	}

	return session, nil
}

func (r *SessionRepository) Create(ctx context.Context, ttl time.Duration, ipAddress, refreshToken string, userID int64) (*entity.Session, *domain.AppError) {
	session := &entity.Session{UserID: userID, IpAddress: ipAddress}

	pipe := r.client.TxPipeline()

	pipe.SAdd(ctx, r.userSessionsKey(userID), refreshToken)
	pipe.Expire(ctx, r.userSessionsKey(userID), ttl)
	pipe.Set(ctx, r.sessionKey(refreshToken), session, ttl)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, domain.ErrUnknownRedis.WithCause(err)
	}

	return session, nil
}

func (r *SessionRepository) Delete(ctx context.Context, refreshToken string, userID int64) (int64, *domain.AppError) {
	pipe := r.client.TxPipeline()

	pipe.SRem(ctx, r.userSessionsKey(userID), refreshToken)
	res := pipe.Del(ctx, r.sessionKey(refreshToken))

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, domain.ErrUnknownRedis.WithCause(err)
	}

	return res.Val(), nil
}

func (r *SessionRepository) DeleteAll(ctx context.Context, userID int64) *domain.AppError {
	pipe := r.client.TxPipeline()

	keys, err := r.client.SMembers(ctx, r.userSessionsKey(userID)).Result()
	if err != nil {
		return domain.ErrUnknownRedis.WithCause(err)
	}

	pipe.Del(ctx, r.userSessionsKey(userID))
	if len(keys) > 0 {
		pipe.Del(ctx, keys...)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return domain.ErrUnknownRedis.WithCause(err)
	}

	return nil
}

func (r *SessionRepository) GetAll(ctx context.Context, userID int64) ([]string, *domain.AppError) {
	keys, err := r.client.SMembers(ctx, r.userSessionsKey(userID)).Result()
	if err != nil {
		return keys, domain.ErrUnknownRedis.WithCause(err)
	}

	return keys, nil
}