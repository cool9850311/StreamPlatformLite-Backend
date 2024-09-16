package usecase

import (
	"Go-Service/src/main/application/usecase"
	"Go-Service/src/main/domain/entity/role"
	"Go-Service/src/main/domain/entity/system"
	"context"
	"errors"
	"testing"

	"Go-Service/src/test/usecase/mock_data"

	"github.com/stretchr/testify/assert"
)

func setup() (*mock_data.MockSystemSettingRepository, *mock_data.MockLogger, usecase.SystemSettingUseCase) {
	mockRepo := new(mock_data.MockSystemSettingRepository)
	mockLogger := new(mock_data.MockLogger)

	useCase := usecase.NewSystemSettingUseCase(mockRepo, mockLogger)
	return mockRepo, mockLogger, *useCase // Dereference the pointer
}

func TestSystemSettingUseCase_GetSetting_AdminUser(t *testing.T) {
	mockRepo, _, useCase := setup()
	ctx := context.Background()

	testSetting := &system.Setting{
		EditorRoleId:        "editor123",
		StreamAccessRoleIds: []string{"user123", "user456"},
	}

	mockRepo.On("GetSetting").Return(testSetting, nil)

	result, err := useCase.GetSetting(ctx, role.Admin)

	assert.Equal(t, testSetting, result)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestSystemSettingUseCase_GetSetting_UnauthorizedUser(t *testing.T) {
	mockRepo, _, useCase := setup()
	ctx := context.Background()

	mockRepo.On("GetSetting").Return(nil, errors.New("unauthorized"))

	result, err := useCase.GetSetting(ctx, role.User)

	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestSystemSettingUseCase_SetSetting_AdminUser(t *testing.T) {
	mockRepo, _, useCase := setup()
	ctx := context.Background()

	testSetting := &system.Setting{
		EditorRoleId:        "editor123",
		StreamAccessRoleIds: []string{"user123", "user456"},
	}

	mockRepo.On("SetSetting", testSetting).Return(nil)

	err := useCase.SetSetting(ctx, testSetting, role.Admin)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestSystemSettingUseCase_SetSetting_UnauthorizedUser(t *testing.T) {
	mockRepo, _, useCase := setup()
	ctx := context.Background()

	testSetting := &system.Setting{
		EditorRoleId:        "editor123",
		StreamAccessRoleIds: []string{"user123", "user456"},
	}

	mockRepo.On("SetSetting", testSetting).Return(errors.New("unauthorized"))

	err := useCase.SetSetting(ctx, testSetting, role.User)

	assert.Error(t, err)
}
