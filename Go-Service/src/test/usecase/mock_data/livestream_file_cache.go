package mock_data

import (
	"github.com/stretchr/testify/mock"
)

type MockFileCache struct {
	mock.Mock
}

func (m *MockFileCache) ReadFile(filePath string) ([]byte, error) {
	args := m.Called(filePath)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockFileCache) StoreCache(filePath string, data []byte) {
	m.Called(filePath, data)
}

func (m *MockFileCache) LoadCache(filePath string) ([]byte, bool) {
	args := m.Called(filePath)
	return args.Get(0).([]byte), args.Bool(1)
}

func (m *MockFileCache) DeleteFile(filePath string) {
	m.Called(filePath)
}

func (m *MockFileCache) Range(f func(key, value interface{}) bool) {
	m.Called(f)
}
