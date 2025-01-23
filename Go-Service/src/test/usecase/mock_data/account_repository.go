package mock_data

import (
	"Go-Service/src/main/domain/entity/account"
	"github.com/stretchr/testify/mock"
)

// MockAccountRepository is a mock implementation of the AccountRepository interface
type MockAccountRepository struct {
	mock.Mock
}

func (m *MockAccountRepository) Create(acc account.Account) error {
	args := m.Called(acc)
	return args.Error(0)
}

func (m *MockAccountRepository) GetAll() ([]account.Account, error) {
	args := m.Called()
	return args.Get(0).([]account.Account), args.Error(1)
}

func (m *MockAccountRepository) GetByUsername(username string) (*account.Account, error) {
	args := m.Called(username)
	return args.Get(0).(*account.Account), args.Error(1)
}

func (m *MockAccountRepository) Update(acc account.Account) error {
	args := m.Called(acc)
	return args.Error(0)
}

func (m *MockAccountRepository) Delete(username string) error {
	args := m.Called(username)
	return args.Error(0)
}
