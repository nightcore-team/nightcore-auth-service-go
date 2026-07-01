package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"nightcore-team/nightcore-auth-service-go/config"
	"nightcore-team/nightcore-auth-service-go/internal/domain/entity"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRepo() (*SessionRepository, redismock.ClientMock) {
	client, mock := redismock.NewClientMock()
	repo := NewSessionRepository(client)
	return repo, mock
}

func TestSessionRepository_Get(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo, mock := newTestRepo()

		session := &entity.Session{UserID: 1, IpAddress: "127.0.0.1"}
		data, err := session.MarshalBinary()
		require.NoError(t, err)

		mock.ExpectGet("session:refresh-token-1").SetVal(string(data))

		got, appErr := repo.Get(context.Background(), "refresh-token-1")

		assert.Nil(t, appErr)
		require.NotNil(t, got)
		assert.Equal(t, session.UserID, got.UserID)
		assert.Equal(t, session.IpAddress, got.IpAddress)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		repo, mock := newTestRepo()

		mock.ExpectGet("session:missing-token").RedisNil()

		got, appErr := repo.Get(context.Background(), "missing-token")

		assert.Nil(t, appErr)
		assert.Nil(t, got)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("redis error", func(t *testing.T) {
		repo, mock := newTestRepo()

		mock.ExpectGet("session:broken-token").SetErr(errors.New("connection refused"))

		got, appErr := repo.Get(context.Background(), "broken-token")

		assert.Nil(t, got)
		require.NotNil(t, appErr)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSessionRepository_GetDel(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo, mock := newTestRepo()

		session := &entity.Session{UserID: 2, IpAddress: "10.0.0.1"}
		data, err := session.MarshalBinary()
		require.NoError(t, err)

		mock.ExpectEval(getDelSessionScript, []string{"session:refresh-token-2"}, "refresh-token-2").
			SetVal(string(data))

		got, appErr := repo.GetDel(context.Background(), "refresh-token-2")

		assert.Nil(t, appErr)
		require.NotNil(t, got)
		assert.Equal(t, session.UserID, got.UserID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		repo, mock := newTestRepo()

		mock.ExpectEval(getDelSessionScript, []string{"session:missing-token"}, "missing-token").
			RedisNil()

		got, appErr := repo.GetDel(context.Background(), "missing-token")

		assert.Nil(t, appErr)
		assert.Nil(t, got)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("redis error", func(t *testing.T) {
		repo, mock := newTestRepo()

		mock.ExpectEval(getDelSessionScript, []string{"session:broken-token"}, "broken-token").
			SetErr(errors.New("connection refused"))

		got, appErr := repo.GetDel(context.Background(), "broken-token")

		assert.Nil(t, got)
		require.NotNil(t, appErr)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSessionRepository_Create(t *testing.T) {
	config.JWT.MaxUserSessions = 10

	repo, mock := newTestRepo()

	userID := int64(3)
	ttl := 24 * time.Hour
	refreshToken := "refresh-token-3"
	ipAddress := "192.168.0.1"

	expectedSession := &entity.Session{UserID: userID, IpAddress: ipAddress}
	data, err := expectedSession.MarshalBinary()
	require.NoError(t, err)

	mock.ExpectEval(createSessionScript,
		[]string{"user_sessions:3", "session:refresh-token-3"},
		refreshToken, int64(ttl.Seconds()), config.JWT.MaxUserSessions, string(data),
	).SetVal(int64(1))

	session, appErr := repo.Create(context.Background(), ttl, ipAddress, refreshToken, userID)

	assert.Nil(t, appErr)
	require.NotNil(t, session)
	assert.Equal(t, userID, session.UserID)
	assert.Equal(t, ipAddress, session.IpAddress)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionRepository_Delete(t *testing.T) {
	repo, mock := newTestRepo()

	userID := int64(4)
	refreshToken := "refresh-token-4"

	mock.ExpectTxPipeline()
	mock.ExpectSRem("user_sessions:4", refreshToken).SetVal(1)
	mock.ExpectDel("session:refresh-token-4").SetVal(1)
	mock.ExpectTxPipelineExec()

	deleted, appErr := repo.Delete(context.Background(), refreshToken, userID)

	assert.Nil(t, appErr)
	assert.Equal(t, int64(1), deleted)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionRepository_DeleteAll(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo, mock := newTestRepo()

		userID := int64(5)

		mock.ExpectEval(deleteAllSessionsScript, []string{"user_sessions:5"}).SetVal(int64(2))

		appErr := repo.DeleteAll(context.Background(), userID)

		assert.Nil(t, appErr)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("redis error", func(t *testing.T) {
		repo, mock := newTestRepo()

		userID := int64(7)

		mock.ExpectEval(deleteAllSessionsScript, []string{"user_sessions:7"}).SetErr(errors.New("timeout"))

		appErr := repo.DeleteAll(context.Background(), userID)

		require.NotNil(t, appErr)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSessionRepository_GetAll(t *testing.T) {
	repo, mock := newTestRepo()

	userID := int64(8)

	mock.ExpectSMembers("user_sessions:8").SetVal([]string{"token-x", "token-y"})

	tokens, appErr := repo.GetAll(context.Background(), userID)

	assert.Nil(t, appErr)
	assert.Equal(t, []string{"token-x", "token-y"}, tokens)
	assert.NoError(t, mock.ExpectationsWereMet())
}
