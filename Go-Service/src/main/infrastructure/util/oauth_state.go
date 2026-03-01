package util

import (
	"Go-Service/src/main/application/interface/state_store"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	stateKeyPrefix = "oauth:state:"
	stateTTL       = 10 * time.Minute
	stateBytes     = 32 // 256 bits
)

type RedisStateStore struct {
	client *redis.Client
}

func NewRedisStateStore(client *redis.Client) state_store.StateStore {
	return &RedisStateStore{client: client}
}

// GenerateState generates a cryptographically secure random state parameter
// and stores it in Redis with a 10-minute expiration
func (s *RedisStateStore) GenerateState(ctx context.Context) (string, error) {
	// Generate random bytes
	b := make([]byte, stateBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	// Base64 URL encode
	state := base64.URLEncoding.EncodeToString(b)

	// Store in Redis with TTL
	key := stateKeyPrefix + state
	err := s.client.Set(ctx, key, "1", stateTTL).Err()
	if err != nil {
		return "", err
	}

	return state, nil
}

// ValidateState validates a state parameter and deletes it from Redis (one-time use)
// Returns error if state is invalid, expired, or already used
func (s *RedisStateStore) ValidateState(ctx context.Context, state string) error {
	if state == "" {
		return errors.New("state is empty")
	}

	key := stateKeyPrefix + state

	// Atomic operation: get and delete
	result, err := s.client.GetDel(ctx, key).Result()
	if err == redis.Nil {
		return errors.New("invalid or expired state")
	}
	if err != nil {
		return err
	}

	if result != "1" {
		return errors.New("invalid state value")
	}

	return nil
}
