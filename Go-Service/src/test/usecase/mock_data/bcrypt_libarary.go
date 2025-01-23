package mock_data

import (
	"github.com/stretchr/testify/mock"
)

// MockBcryptGenerator is a mock implementation of the BcryptGenerator interface
type MockBcryptGenerator struct {
	mock.Mock
}

// HashPassword is a mock implementation of the HashPassword method
func (m *MockBcryptGenerator) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

// CheckPasswordHash is a mock implementation of the CheckPasswordHash method
func (m *MockBcryptGenerator) CheckPasswordHash(password, hash string) bool {
	args := m.Called(password, hash)
	return args.Bool(0)
}
