package mock_data

import (
	"Go-Service/src/main/domain/entity/livestream"
	"github.com/stretchr/testify/mock"
)

type MockLivestreamRepository struct {
	mock.Mock
}

func (m *MockLivestreamRepository) GetByID(id string) (*livestream.Livestream, error) {
	args := m.Called(id)
	if args.Get(0) != nil {
		return args.Get(0).(*livestream.Livestream), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockLivestreamRepository) GetByOwnerID(ownerID string) (*livestream.Livestream, error) {
	args := m.Called(ownerID)
	if args.Get(0) != nil {
		return args.Get(0).(*livestream.Livestream), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockLivestreamRepository) GetOne() (*livestream.Livestream, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).(*livestream.Livestream), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockLivestreamRepository) Create(livestream *livestream.Livestream) error {
	args := m.Called(livestream)
	return args.Error(0)
}

func (m *MockLivestreamRepository) Update(livestream *livestream.Livestream) error {
	args := m.Called(livestream)
	return args.Error(0)
}

func (m *MockLivestreamRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}
func (m *MockLivestreamRepository) MuteUser(livestreamUUID string, userID string) error {
	args := m.Called(livestreamUUID, userID)
	return args.Error(0)
}