package cache

import (
	"Go-Service/src/main/domain/entity/chat"
)

type Chat interface {
	GetChat(livestreamUUID string, index string, count int) ([]chat.Chat, error)
	AddChat(livestreamUUID string, chat chat.Chat) error
	DeleteChat(livestreamUUID string, chatID string) error
}