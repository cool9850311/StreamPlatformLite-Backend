package mock_data

import (
	"github.com/stretchr/testify/mock"
)

type MockLivestreamService struct {
	mock.Mock
}

func (m *MockLivestreamService) OpenStream(name, uuid, apiKey, outputPathUUID string) error {
	return nil
}

func (m *MockLivestreamService) UpdateStreamOutPutPathUUID(uuid, outputPathUUID string) error {
	return nil
}

func (m *MockLivestreamService) CloseStream(uuid string) error {
	return nil
}

func (m *MockLivestreamService) StartService() error {
	return nil
}

func (m *MockLivestreamService) RunLoop() error {
	return nil
}

func (m *MockLivestreamService) IsLiveStreamExist(uuid string) bool {
	return true
}
