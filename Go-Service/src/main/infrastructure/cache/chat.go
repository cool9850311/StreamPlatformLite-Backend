package cache

import (
	"Go-Service/src/main/application/interface/cache"
	"Go-Service/src/main/domain/entity/chat"
	"context"

	"github.com/redis/go-redis/v9"
)

type RedisChat struct {
	client *redis.Client
}

func NewRedisChat(client *redis.Client) cache.Chat {
	return &RedisChat{client: client}
}

func (r *RedisChat) GetChat(livestreamUUID string, index string, count int) ([]chat.Chat, error) {
	ctx := context.Background()
	key := "chat_" + livestreamUUID

	var streams []redis.XMessage
	var err error
	if index == "-1" {
		// Fetch the newest messages
		streams, err = r.client.XRevRangeN(ctx, key, "+", "-", int64(count)).Result()
		// Reverse the order of streams
		for i, j := 0, len(streams)-1; i < j; i, j = i+1, j-1 {
			streams[i], streams[j] = streams[j], streams[i]
		}
	} else {
		// Fetch messages after the given index
		streams, err = r.client.XRangeN(ctx, key, "("+index, "+", int64(count)).Result()
	}
	// if index != "-1" {
	// 	// Fetch previous messages from the given index
	// 	streams, err = r.client.XRevRangeN(ctx, key, index, "-", int64(count)).Result()
	// } else {
	// 	// Fetch the newest messages
	// 	streams, err = r.client.XRangeN(ctx, key, "-", "+", int64(count)).Result()
	// }

	if err != nil {
		return nil, err
	}

	chats := make([]chat.Chat, 0, len(streams))
	for _, stream := range streams {
		chats = append(chats, chat.Chat{
			ID:       stream.ID,
			UserID:   stream.Values["user_id"].(string),
			Username: stream.Values["username"].(string),
			Message:  stream.Values["message"].(string),
		})
	}

	return chats, nil
}

func (r *RedisChat) AddChat(livestreamUUID string, chat chat.Chat) error {
	ctx := context.Background()
	key := "chat_" + livestreamUUID

	// Add message to the stream
	_, err := r.client.XAdd(ctx, &redis.XAddArgs{
		Stream: key,
		Values: map[string]interface{}{
			"user_id":  chat.UserID,
			"username": chat.Username,
			"message":  chat.Message,
		},
	}).Result()

	return err
}

func (r *RedisChat) DeleteChat(livestreamUUID string, chatID string) error {
	ctx := context.Background()
	key := "chat_" + livestreamUUID

	// Delete message from the stream
	_, err := r.client.XDel(ctx, key, chatID).Result()
	return err
}
