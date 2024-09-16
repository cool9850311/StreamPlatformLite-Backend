package mock_data

import (
	"Go-Service/src/main/domain/entity/system"
	"github.com/stretchr/testify/mock"
)

type MockSystemSettingRepository struct {
	mock.Mock
}

func (m *MockSystemSettingRepository) GetSetting() (*system.Setting, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).(*system.Setting), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockSystemSettingRepository) SetSetting(setting *system.Setting) error {
	args := m.Called(setting)
	return args.Error(0)
}