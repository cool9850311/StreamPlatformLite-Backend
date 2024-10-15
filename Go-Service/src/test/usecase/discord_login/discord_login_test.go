package usecase

import (
	"Go-Service/src/main/application/dto"
	"Go-Service/src/main/application/dto/config"
	"Go-Service/src/main/application/usecase"
	"Go-Service/src/main/domain/entity/role"
	"Go-Service/src/main/domain/entity/system"
	"Go-Service/src/test/usecase/mock_data"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setup() (*mock_data.MockSystemSettingRepository, *mock_data.MockLogger, usecase.DiscordLoginUseCase, *mock_data.MockDiscordOAuth, *mock_data.MockJWTGenerator) {
	mockRepo := new(mock_data.MockSystemSettingRepository)
	mockLogger := new(mock_data.MockLogger)
	mockDiscordOAuth := new(mock_data.MockDiscordOAuth)
	mockJWTGenerator := new(mock_data.MockJWTGenerator)
	cfg := config.Config{
		Discord: struct {
			ClientID     string `mapstructure:"clientId"`
			ClientSecret string `mapstructure:"clientSecret"`
			AdminID      string `mapstructure:"adminId"`
			GuildID      string `mapstructure:"guildId"`
		}{
			ClientID:     "fakeClientID",
			ClientSecret: "fakeClientSecret",
			AdminID:      "admin123",
			GuildID:      "fakeGuildID",
		},
		Server: struct {
			Port         int    `mapstructure:"port"`
			Domain       string `mapstructure:"domain"`
			HTTPS        bool   `mapstructure:"https" default:"false"`
			EnableGinLog bool   `mapstructure:"enable_gin_log" default:"true"`
		}{
			Port:         8080, // or any appropriate test value
			Domain:       "fakeServerDomain",
			HTTPS:        false, // or any appropriate test value
			EnableGinLog: true,  // or any appropriate test value
		},
		Frontend: struct {
			Domain string `mapstructure:"domain"`
			Port   int    `mapstructure:"port"`
		}{
			Domain: "fakeFrontendDomain",
			Port:   3000, // or any appropriate default value
		},
		JWT: struct {
			SecretKey string `mapstructure:"secretKey"`
		}{
			SecretKey: "fakeJWTSecret",
		},
	}

	useCase := usecase.NewDiscordLoginUseCase(mockRepo, mockLogger, cfg, mockDiscordOAuth, mockJWTGenerator)
	return mockRepo, mockLogger, *useCase, mockDiscordOAuth, mockJWTGenerator // Dereference the pointer
}

func TestDiscordLoginUseCase_AdminUser(t *testing.T) {
	mockRepo, _, useCase, mockDiscordOAuth, mockJWTGenerator := setup()
	ctx := context.Background()

	testSetting := &system.Setting{
		EditorRoleId:        "editor123",
		StreamAccessRoleIds: []string{"user123", "user456"},
	}

	mockRepo.On("GetSetting").Return(testSetting, nil)
	mockDiscordOAuth.On("GetAccessToken", ctx, "fakeClientID", "fakeClientSecret", "adminCode", "http://fakeServerDomain:8080/oauth/discord").Return("accessToken", nil)
	mockDiscordOAuth.On("GetGuildMemberData", ctx, "accessToken", "fakeGuildID").Return(&dto.DiscordGuildMemberDTO{
		User:  dto.DiscordUserDTO{ID: "admin123"},
		Roles: []string{},
	}, nil)
	mockJWTGenerator.On("GenerateToken", ctx, "admin123", mock.AnythingOfType("*dto.DiscordGuildMemberDTO"), role.Admin, "fakeJWTSecret").Return("jwtToken", nil)
	result, err := useCase.Login(ctx, "adminCode")

	assert.Contains(t, result, "token=")
	assert.NoError(t, err)
}

func TestDiscordLoginUseCase_EditorRole(t *testing.T) {
	mockRepo, _, useCase, mockDiscordOAuth, mockJWTGenerator := setup()
	ctx := context.Background()

	testSetting := &system.Setting{
		EditorRoleId:        "editor123",
		StreamAccessRoleIds: []string{"user123", "user456"},
	}
	mockRepo.On("GetSetting").Return(testSetting, nil)
	mockDiscordOAuth.On("GetAccessToken", ctx, "fakeClientID", "fakeClientSecret", "editorCode", "http://fakeServerDomain:8080/oauth/discord").Return("accessToken", nil)
	mockDiscordOAuth.On("GetGuildMemberData", ctx, "accessToken", "fakeGuildID").Return(&dto.DiscordGuildMemberDTO{
		User:  dto.DiscordUserDTO{ID: "fakeEditorID"},
		Roles: []string{"editor123"},
	}, nil)
	mockJWTGenerator.On("GenerateToken", ctx, "fakeEditorID", mock.AnythingOfType("*dto.DiscordGuildMemberDTO"), role.Editor, "fakeJWTSecret").Return("jwtToken", nil)
	result, err := useCase.Login(ctx, "editorCode")

	assert.Contains(t, result, "token=")
	assert.NoError(t, err)
}

func TestDiscordLoginUseCase_UserRoleWithStreamAccess(t *testing.T) {
	mockRepo, _, useCase, mockDiscordOAuth, mockJWTGenerator := setup()
	ctx := context.Background()

	testSetting := &system.Setting{
		EditorRoleId:        "editor123",
		StreamAccessRoleIds: []string{"user123", "user456"},
	}
	mockRepo.On("GetSetting").Return(testSetting, nil)
	mockDiscordOAuth.On("GetAccessToken", ctx, "fakeClientID", "fakeClientSecret", "userCode", "http://fakeServerDomain:8080/oauth/discord").Return("accessToken", nil)
	mockDiscordOAuth.On("GetGuildMemberData", ctx, "accessToken", "fakeGuildID").Return(&dto.DiscordGuildMemberDTO{
		User:  dto.DiscordUserDTO{ID: "fakeUserID"},
		Roles: []string{"user123"},
	}, nil)
	mockJWTGenerator.On("GenerateToken", ctx, "fakeUserID", mock.AnythingOfType("*dto.DiscordGuildMemberDTO"), role.User, "fakeJWTSecret").Return("jwtToken", nil)
	result, err := useCase.Login(ctx, "userCode")

	assert.Contains(t, result, "token=")
	assert.NoError(t, err)
}

func TestDiscordLoginUseCase_GuestRole(t *testing.T) {
	mockRepo, _, useCase, mockDiscordOAuth, mockJWTGenerator := setup()
	ctx := context.Background()

	testSetting := &system.Setting{
		EditorRoleId:        "editor123",
		StreamAccessRoleIds: []string{"user123", "user456"},
	}
	mockRepo.On("GetSetting").Return(testSetting, nil)
	mockDiscordOAuth.On("GetAccessToken", ctx, "fakeClientID", "fakeClientSecret", "guestCode", "http://fakeServerDomain:8080/oauth/discord").Return("accessToken", nil)
	mockDiscordOAuth.On("GetGuildMemberData", ctx, "accessToken", "fakeGuildID").Return(&dto.DiscordGuildMemberDTO{
		User:  dto.DiscordUserDTO{ID: "fakeGuestID"},
		Roles: []string{},
	}, nil)
	mockJWTGenerator.On("GenerateToken", ctx, "fakeGuestID", mock.AnythingOfType("*dto.DiscordGuildMemberDTO"), role.Guest, "fakeJWTSecret").Return("jwtToken", nil)
	result, err := useCase.Login(ctx, "guestCode")

	assert.Contains(t, result, "token=")
	assert.NoError(t, err)
}

func TestDiscordLoginUseCase_SystemSettingRetrievalError(t *testing.T) {
	mockRepo, _, useCase, mockDiscordOAuth, _ := setup()
	ctx := context.Background()

	mockRepo.On("GetSetting").Return(nil, errors.New("database error"))
	mockDiscordOAuth.On("GetAccessToken", ctx, "fakeClientID", "fakeClientSecret", "errorCode", "http://fakeServerDomain:8080/oauth/discord").Return("accessToken", nil)
	mockDiscordOAuth.On("GetGuildMemberData", ctx, "accessToken", "fakeGuildID").Return(&dto.DiscordGuildMemberDTO{
		User:  dto.DiscordUserDTO{ID: "user999"},
		Roles: []string{},
	}, nil)

	result, err := useCase.Login(ctx, "errorCode")

	assert.Equal(t, "http://fakeFrontendDomain:3000", result)
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}
