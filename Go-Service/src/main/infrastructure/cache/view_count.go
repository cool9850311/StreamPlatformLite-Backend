package cache

import (
	"Go-Service/src/main/application/interface/cache"
	"context"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

type RedisViewerCount struct {
	client *redis.Client
}

func NewRedisViewerCount(client *redis.Client) cache.ViewerCount {
	return &RedisViewerCount{client: client}
}

func (r *RedisViewerCount) GetViewerCount(livestreamUUID string) (int, error) {
	ctx := context.Background()
	key := "viewer_count_" + livestreamUUID
	count, err := r.client.ZCard(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *RedisViewerCount) AddViewerCount(livestreamUUID string, userID string) error {
	ctx := context.Background()
	key := "viewer_count_" + livestreamUUID
	score := float64(time.Now().Unix())
	_, err := r.client.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: userID,
	}).Result()
	return err
}
func (r *RedisViewerCount) RemoveViewerCount(livestreamUUID string, seconds int) (int, error) {
	ctx := context.Background()
	key := "viewer_count_" + livestreamUUID
	now := time.Now().Unix() - int64(seconds)

	// Remove elements with scores less than 'now'
	_, err := r.client.ZRemRangeByScore(ctx, key, "-inf", strconv.FormatInt(now, 10)).Result()
	if err != nil {
		return 0, err
	}

	// Get the updated count after removal
	count, err := r.client.ZCard(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	return int(count), nil

}
