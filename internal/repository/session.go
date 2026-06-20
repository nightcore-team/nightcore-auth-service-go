package repository

import (
	"fmt"
	"nightcore-team/nightcore-auth-service-go/internal/domain/entity"
	"time"

	"github.com/go-redis/redis"
)

func NewSessionRepository(redisClient *redis.Client) *SessionRepository {
	return &SessionRepository{
		client: redisClient,
	}
}

type SessionRepository struct {
	client *redis.Client
}

func (r *SessionRepository) sessionKey(refreshToken string) string {
	return fmt.Sprintf("session:%s", refreshToken)
}

func (r *SessionRepository) userSessionsKey(userID int64) string {
	return fmt.Sprintf("user_sessions:%d", userID)
}

func (r *SessionRepository) Get(refreshToken string) *entity.Session {
	cmd := r.client.Get(r.sessionKey(refreshToken))

	if cmd == nil {
		return nil
	}

	session := &entity.Session{}

	err := cmd.Scan(session)
	if err != nil {
		panic("")
	}

	return session
}

func (r *SessionRepository) Create(ttl int64, ipAddress, refreshToken string, userID int64) *entity.Session {
	session := &entity.Session{UserID: userID, IpAddress: ipAddress}

	pipe := r.client.TxPipeline()

	pipe.SAdd(r.userSessionsKey(userID), refreshToken)
	pipe.Expire(r.userSessionsKey(userID), time.Duration(ttl))
	pipe.Set(r.sessionKey(refreshToken), &session, time.Duration(ttl))

	_, err := pipe.Exec()
	if err != nil {
		panic("")
	}

	return session
}

func (r *SessionRepository) Delete(refreshToken string, userID int64) int64 {
	pipe := r.client.TxPipeline()

	pipe.SRem(r.userSessionsKey(userID), refreshToken)
	res := pipe.Del(r.sessionKey(refreshToken))

	_, err := pipe.Exec()
	if err != nil {
		panic("")
	}

	return res.Val()
}

func (r *SessionRepository) DeleteAll(userID int64) {
	keys := []string{}

	pipe := r.client.TxPipeline()

	cmd := r.client.Get(r.userSessionsKey(userID))
	err := cmd.Scan(&keys)
	if err != nil {
		panic("")
	}

	pipe.Del(r.userSessionsKey(userID))
	pipe.Del(keys...)

	_, err = pipe.Exec()
	if err != nil {
		panic("")
	}
}

func (r *SessionRepository) GetAll(userID int64) []string {
	keys := []string{}

	cmd := r.client.Get(r.userSessionsKey(userID))
	err := cmd.Scan(&keys)
	if err != nil {
		panic("")
	}

	return keys
}