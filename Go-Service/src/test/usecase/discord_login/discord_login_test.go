package usecase

import (
	"Go-Service/src/main/application/dto/config"
	"Go-Service/src/main/application/usecase"
	"Go-Service/src/main/domain/entity/role"
	"Go-Service/src/main/domain/entity/system"
	"context"
	"errors"
	"testing"

	"Go-Service/src/test/usecase/mock_data"

	"github.com/stretchr/testify/assert"
)

func setup() (*mock_data.MockSystemSettingRepository, *mock_data.MockLogger, usecase.DiscordLoginUseCase) {
	mockRepo := new(mock_data.MockSystemSettingRepository)
	mockLogger := new(mock_data.MockLogger)

	cfg := config.Config{
		Discord: struct {
			ClientID     string `mapstructure:"clientId"`
			ClientSecret string `mapstructure:"clientSecret"`
			AdminID      string `mapstructure:"adminId"`
			GuildID      string `mapstructure:"guildId"`
		}{
			AdminID: "admin123",
		},
	}

	useCase := usecase.NewDiscordLoginUseCase(mockRepo, mockLogger, cfg)
	return mockRepo, mockLogger, *useCase // Dereference the pointer
}

func TestDiscordLoginUseCase_AdminUser(t *testing.T) {
	mockRepo, _, useCase := setup()
	ctx := context.Background()

	testSetting := &system.Setting{
		EditorRoleId:        "editor123",
		StreamAccessRoleIds: []string{"user123", "user456"},
	}

	mockRepo.On("GetSetting").Return(testSetting, nil)

	result, err := useCase.Login(ctx, "admin123", []string{})

	assert.Equal(t, role.Admin, result)
	assert.NoError(t, err)
}

func TestDiscordLoginUseCase_EditorRole(t *testing.T) {
	mockRepo, _, useCase := setup()
	ctx := context.Background()

	testSetting := &system.Setting{
		EditorRoleId:        "editor123",
		StreamAccessRoleIds: []string{"user123", "user456"},
	}

	mockRepo.On("GetSetting").Return(testSetting, nil)

	result, err := useCase.Login(ctx, "user789", []string{"editor123"})

	assert.Equal(t, role.Editor, result)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDiscordLoginUseCase_UserRoleWithStreamAccess(t *testing.T) {
	mockRepo, _, useCase := setup()
	ctx := context.Background()

	testSetting := &system.Setting{
		EditorRoleId:        "editor123",
		StreamAccessRoleIds: []string{"user123", "user456"},
	}

	mockRepo.On("GetSetting").Return(testSetting, nil)

	result, err := useCase.Login(ctx, "user123", []string{"user123"})

	assert.Equal(t, role.User, result)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDiscordLoginUseCase_GuestRole(t *testing.T) {
	mockRepo, _, useCase := setup()
	ctx := context.Background()

	testSetting := &system.Setting{
		EditorRoleId:        "editor123",
		StreamAccessRoleIds: []string{"user123", "user456"},
	}

	mockRepo.On("GetSetting").Return(testSetting, nil)

	result, err := useCase.Login(ctx, "user999", []string{"unknownRole"})

	assert.Equal(t, role.Guest, result)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDiscordLoginUseCase_SystemSettingRetrievalError(t *testing.T) {
	mockRepo, _, useCase := setup()
	ctx := context.Background()

	mockRepo.On("GetSetting").Return(nil, errors.New("database error"))

	result, err := useCase.Login(ctx, "user999", []string{})

	assert.Equal(t, role.Guest, result)
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}
