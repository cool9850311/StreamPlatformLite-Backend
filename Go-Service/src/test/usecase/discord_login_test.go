package usecase

import (
	"context"
	"errors"
	"Go-Service/src/main/application/dto/config"
	"Go-Service/src/main/application/usecase"
	"Go-Service/src/main/domain/entity/role"
	"Go-Service/src/main/domain/entity/system"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSystemSettingRepository is a mock implementation of SystemSettingRepository
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

type MockLogger struct{}

func (m *MockLogger) Panic(ctx context.Context, msg string) {}
func (m *MockLogger) Fatal(ctx context.Context, msg string) {}
func (m *MockLogger) Error(ctx context.Context, msg string) {}
func (m *MockLogger) Warn(ctx context.Context, msg string)  {}
func (m *MockLogger) Info(ctx context.Context, msg string)  {}
func (m *MockLogger) Debug(ctx context.Context, msg string) {}
func (m *MockLogger) Trace(ctx context.Context, msg string) {}

func setup() (*MockSystemSettingRepository, *MockLogger, usecase.DiscordLoginUseCase) {
	mockRepo := new(MockSystemSettingRepository)
	mockLogger := new(MockLogger)

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
	return mockRepo, mockLogger, *useCase  // Dereference the pointer
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
