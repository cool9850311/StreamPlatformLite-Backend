package mock_data

import (
	"Go-Service/src/main/application/interface/state_store"
	"context"

	"github.com/stretchr/testify/mock"
)

var _ state_store.StateStore = (*MockStateStore)(nil)

type MockStateStore struct {
	mock.Mock
}

// Mock implementation of GenerateState
func (m *MockStateStore) GenerateState(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

// Mock implementation of ValidateState
func (m *MockStateStore) ValidateState(ctx context.Context, state string) error {
	args := m.Called(ctx, state)
	return args.Error(0)
}
