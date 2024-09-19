package mock_data

import (
	"github.com/stretchr/testify/mock"
)

type MockViewerCountCache struct {
	mock.Mock
}

func (m *MockViewerCountCache) GetViewerCount(livestreamUUID string) (int, error) {
	args := m.Called(livestreamUUID)
	return args.Int(0), args.Error(1)
}

func (m *MockViewerCountCache) AddViewerCount(livestreamUUID string, userID string) error {
	args := m.Called(livestreamUUID, userID)
	return args.Error(0)
}

func (m *MockViewerCountCache) RemoveViewerCount(livestreamUUID string, seconds int) (int, error) {
	args := m.Called(livestreamUUID, seconds)
	return args.Int(0), args.Error(1)
}
