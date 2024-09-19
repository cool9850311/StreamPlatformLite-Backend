package mock_data

import (
	"Go-Service/src/main/domain/entity/chat"

	"github.com/stretchr/testify/mock"
)

type MockChatCache struct {
	mock.Mock
}

func (m *MockChatCache) GetChat(livestreamUUID string, index string, count int) ([]chat.Chat, error) {
	args := m.Called(livestreamUUID, index, count)
	return args.Get(0).([]chat.Chat), args.Error(1)
}

func (m *MockChatCache) AddChat(livestreamUUID string, chat chat.Chat) error {
	args := m.Called(livestreamUUID, chat)
	return args.Error(0)
}

func (m *MockChatCache) DeleteChat(livestreamUUID string, chatID string) error {
	args := m.Called(livestreamUUID, chatID)
	return args.Error(0)
}

func (m *MockChatCache) GetDeleteChatIDs(livestreamUUID string) ([]string, error) {
	args := m.Called(livestreamUUID)
	return args.Get(0).([]string), args.Error(1)
}