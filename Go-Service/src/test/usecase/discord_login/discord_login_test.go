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

func setup() (*mock_data.MockSystemSettingRepository, *mock_data.MockLogger, usecase.DiscordLoginUseCase, *mock_data.MockDiscordOAuth, *mock_data.MockJWTGenerator, *mock_data.MockStateStore) {
	mockRepo := new(mock_data.MockSystemSettingRepository)
	mockLogger := new(mock_data.MockLogger)
	mockDiscordOAuth := new(mock_data.MockDiscordOAuth)
	mockJWTGenerator := new(mock_data.MockJWTGenerator)
	mockStateStore := new(mock_data.MockStateStore)
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

	useCase := usecase.NewDiscordLoginUseCase(mockRepo, mockLogger, cfg, mockDiscordOAuth, mockJWTGenerator, mockStateStore)
	return mockRepo, mockLogger, *useCase, mockDiscordOAuth, mockJWTGenerator, mockStateStore // Dereference the pointer
}

func TestDiscordLoginUseCase_AdminUser(t *testing.T) {
	mockRepo, _, useCase, mockDiscordOAuth, mockJWTGenerator, _ := setup()
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
	mockJWTGenerator.On("GenerateDiscordToken", ctx, "admin123", mock.AnythingOfType("*dto.DiscordGuildMemberDTO"), role.Admin, "fakeJWTSecret").Return("jwtToken", nil)
	token, redirectURL, err := useCase.Login(ctx, "adminCode")

	assert.NotEmpty(t, token)
	assert.Equal(t, "http://fakeFrontendDomain:3000/stream", redirectURL)
	assert.NoError(t, err)
}

func TestDiscordLoginUseCase_EditorRole(t *testing.T) {
	mockRepo, _, useCase, mockDiscordOAuth, mockJWTGenerator, _ := setup()
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
	mockJWTGenerator.On("GenerateDiscordToken", ctx, "fakeEditorID", mock.AnythingOfType("*dto.DiscordGuildMemberDTO"), role.Editor, "fakeJWTSecret").Return("jwtToken", nil)
	token, redirectURL, err := useCase.Login(ctx, "editorCode")

	assert.NotEmpty(t, token)
	assert.Equal(t, "http://fakeFrontendDomain:3000/stream", redirectURL)
	assert.NoError(t, err)
}

func TestDiscordLoginUseCase_UserRoleWithStreamAccess(t *testing.T) {
	mockRepo, _, useCase, mockDiscordOAuth, mockJWTGenerator, _ := setup()
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
	mockJWTGenerator.On("GenerateDiscordToken", ctx, "fakeUserID", mock.AnythingOfType("*dto.DiscordGuildMemberDTO"), role.User, "fakeJWTSecret").Return("jwtToken", nil)
	token, redirectURL, err := useCase.Login(ctx, "userCode")

	assert.NotEmpty(t, token)
	assert.Equal(t, "http://fakeFrontendDomain:3000/stream", redirectURL)
	assert.NoError(t, err)
}

func TestDiscordLoginUseCase_GuestRole(t *testing.T) {
	mockRepo, _, useCase, mockDiscordOAuth, mockJWTGenerator, _ := setup()
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
	mockJWTGenerator.On("GenerateDiscordToken", ctx, "fakeGuestID", mock.AnythingOfType("*dto.DiscordGuildMemberDTO"), role.Guest, "fakeJWTSecret").Return("jwtToken", nil)
	token, redirectURL, err := useCase.Login(ctx, "guestCode")

	assert.NotEmpty(t, token)
	assert.Equal(t, "http://fakeFrontendDomain:3000/stream", redirectURL)
	assert.NoError(t, err)
}

func TestDiscordLoginUseCase_SystemSettingRetrievalError(t *testing.T) {
	mockRepo, _, useCase, mockDiscordOAuth, _, _ := setup()
	ctx := context.Background()

	mockRepo.On("GetSetting").Return(nil, errors.New("database error"))
	mockDiscordOAuth.On("GetAccessToken", ctx, "fakeClientID", "fakeClientSecret", "errorCode", "http://fakeServerDomain:8080/oauth/discord").Return("accessToken", nil)
	mockDiscordOAuth.On("GetGuildMemberData", ctx, "accessToken", "fakeGuildID").Return(&dto.DiscordGuildMemberDTO{
		User:  dto.DiscordUserDTO{ID: "user999"},
		Roles: []string{},
	}, nil)

	token, redirectURL, err := useCase.Login(ctx, "errorCode")

	assert.Empty(t, token)
	assert.Equal(t, "http://fakeFrontendDomain:3000/stream", redirectURL)
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDiscordLoginUseCase_InitiateLogin_Success(t *testing.T) {
	_, _, useCase, _, _, mockStateStore := setup()
	ctx := context.Background()

	// Mock state generation
	mockStateStore.On("GenerateState", ctx).Return("test_state_123", nil)

	authURL, err := useCase.InitiateLogin(ctx)

	assert.NoError(t, err)
	assert.Contains(t, authURL, "https://discord.com/api/oauth2/authorize")
	assert.Contains(t, authURL, "client_id=fakeClientID")
	assert.Contains(t, authURL, "redirect_uri=http%3A%2F%2FfakeServerDomain%3A8080%2Foauth%2Fdiscord")
	assert.Contains(t, authURL, "state=test_state_123")
	assert.Contains(t, authURL, "scope=identify+email+guilds.members.read")
	mockStateStore.AssertExpectations(t)
}

func TestDiscordLoginUseCase_InitiateLogin_StateGenerationError(t *testing.T) {
	_, _, useCase, _, _, mockStateStore := setup()
	ctx := context.Background()

	// Mock state generation failure
	mockStateStore.On("GenerateState", ctx).Return("", errors.New("redis error"))

	authURL, err := useCase.InitiateLogin(ctx)

	assert.Error(t, err)
	assert.Empty(t, authURL)
	mockStateStore.AssertExpectations(t)
}

func TestDiscordLoginUseCase_ValidateStateAndLogin_Success(t *testing.T) {
	mockRepo, _, useCase, mockDiscordOAuth, mockJWTGenerator, mockStateStore := setup()
	ctx := context.Background()

	testSetting := &system.Setting{
		EditorRoleId:        "editor123",
		StreamAccessRoleIds: []string{"user123", "user456"},
	}

	// Mock state validation success
	mockStateStore.On("ValidateState", ctx, "valid_state").Return(nil)
	mockRepo.On("GetSetting").Return(testSetting, nil)
	mockDiscordOAuth.On("GetAccessToken", ctx, "fakeClientID", "fakeClientSecret", "test_code", "http://fakeServerDomain:8080/oauth/discord").Return("accessToken", nil)
	mockDiscordOAuth.On("GetGuildMemberData", ctx, "accessToken", "fakeGuildID").Return(&dto.DiscordGuildMemberDTO{
		User:  dto.DiscordUserDTO{ID: "admin123"},
		Roles: []string{},
	}, nil)
	mockJWTGenerator.On("GenerateDiscordToken", ctx, "admin123", mock.AnythingOfType("*dto.DiscordGuildMemberDTO"), role.Admin, "fakeJWTSecret").Return("jwtToken", nil)

	token, successURL, errorURL, err := useCase.ValidateStateAndLogin(ctx, "test_code", "valid_state")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, "http://fakeFrontendDomain:3000/stream", successURL)
	assert.Equal(t, "http://fakeFrontendDomain:3000/?error=invalid_state", errorURL)
	mockStateStore.AssertExpectations(t)
}

func TestDiscordLoginUseCase_ValidateStateAndLogin_InvalidState(t *testing.T) {
	_, _, useCase, _, _, mockStateStore := setup()
	ctx := context.Background()

	// Mock state validation failure
	mockStateStore.On("ValidateState", ctx, "invalid_state").Return(errors.New("invalid state"))

	token, successURL, errorURL, err := useCase.ValidateStateAndLogin(ctx, "test_code", "invalid_state")

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Empty(t, successURL)
	assert.Equal(t, "http://fakeFrontendDomain:3000/?error=invalid_state", errorURL)
	mockStateStore.AssertExpectations(t)
}

// ==================== 4.1 单元测试场景 ====================

// Test: State 重放攻击防护
func TestDiscordLoginUseCase_ValidateStateAndLogin_ReplayAttack(t *testing.T) {
	mockRepo, _, useCase, mockDiscordOAuth, mockJWTGenerator, mockStateStore := setup()
	ctx := context.Background()

	testSetting := &system.Setting{
		EditorRoleId:        "editor123",
		StreamAccessRoleIds: []string{"user123", "user456"},
	}

	state := "replay_state_123"

	// First validation: Success
	mockStateStore.On("ValidateState", ctx, state).Return(nil).Once()
	mockRepo.On("GetSetting").Return(testSetting, nil).Once()
	mockDiscordOAuth.On("GetAccessToken", ctx, "fakeClientID", "fakeClientSecret", "code1", "http://fakeServerDomain:8080/oauth/discord").Return("accessToken", nil).Once()
	mockDiscordOAuth.On("GetGuildMemberData", ctx, "accessToken", "fakeGuildID").Return(&dto.DiscordGuildMemberDTO{
		User:  dto.DiscordUserDTO{ID: "admin123"},
		Roles: []string{},
	}, nil).Once()
	mockJWTGenerator.On("GenerateDiscordToken", ctx, "admin123", mock.AnythingOfType("*dto.DiscordGuildMemberDTO"), role.Admin, "fakeJWTSecret").Return("jwtToken", nil).Once()

	// First attempt: Should succeed
	token1, _, _, err1 := useCase.ValidateStateAndLogin(ctx, "code1", state)
	assert.NoError(t, err1)
	assert.NotEmpty(t, token1)

	// Second validation: Should fail (state already used/deleted)
	mockStateStore.On("ValidateState", ctx, state).Return(errors.New("invalid or expired state")).Once()

	// Second attempt: Should fail (replay attack)
	token2, successURL2, errorURL2, err2 := useCase.ValidateStateAndLogin(ctx, "code2", state)
	assert.Error(t, err2)
	assert.Empty(t, token2)
	assert.Empty(t, successURL2)
	assert.Equal(t, "http://fakeFrontendDomain:3000/?error=invalid_state", errorURL2)

	mockStateStore.AssertExpectations(t)
}

// ==================== 4.2 集成测试场景（转为 UseCase 测试）====================

// Test: 完整正常登录流程（InitiateLogin → ValidateStateAndLogin）
func TestDiscordLoginUseCase_CompleteOAuthFlow_NormalCase(t *testing.T) {
	_, _, useCase, mockDiscordOAuth, mockJWTGenerator, mockStateStore := setup()
	ctx := context.Background()

	generatedState := "complete_flow_state_xyz"

	// Step 1: Initiate login (generate state)
	mockStateStore.On("GenerateState", ctx).Return(generatedState, nil).Once()

	authURL, err := useCase.InitiateLogin(ctx)
	assert.NoError(t, err)
	assert.Contains(t, authURL, "state="+generatedState)
	assert.Contains(t, authURL, "https://discord.com/api/oauth2/authorize")

	// Step 2: Validate state and login (Admin user - no GetSetting needed)
	mockStateStore.On("ValidateState", ctx, generatedState).Return(nil).Once()
	mockDiscordOAuth.On("GetAccessToken", ctx, "fakeClientID", "fakeClientSecret", "auth_code_123", "http://fakeServerDomain:8080/oauth/discord").Return("discord_access_token", nil).Once()
	mockDiscordOAuth.On("GetGuildMemberData", ctx, "discord_access_token", "fakeGuildID").Return(&dto.DiscordGuildMemberDTO{
		User:  dto.DiscordUserDTO{ID: "admin123"}, // Admin user
		Roles: []string{},
	}, nil).Once()
	mockJWTGenerator.On("GenerateDiscordToken", ctx, "admin123", mock.AnythingOfType("*dto.DiscordGuildMemberDTO"), role.Admin, "fakeJWTSecret").Return("final_jwt_token", nil).Once()

	token, successURL, _, err := useCase.ValidateStateAndLogin(ctx, "auth_code_123", generatedState)

	assert.NoError(t, err)
	assert.Equal(t, "final_jwt_token", token)
	assert.Equal(t, "http://fakeFrontendDomain:3000/stream", successURL)

	mockStateStore.AssertExpectations(t)
	mockDiscordOAuth.AssertExpectations(t)
	mockJWTGenerator.AssertExpectations(t)
}

// Test: Code 缺失场景
func TestDiscordLoginUseCase_ValidateStateAndLogin_MissingCode(t *testing.T) {
	mockRepo, _, useCase, mockDiscordOAuth, _, mockStateStore := setup()
	ctx := context.Background()

	// State validation succeeds, but code is empty
	mockStateStore.On("ValidateState", ctx, "valid_state").Return(nil).Once()
	mockDiscordOAuth.On("GetAccessToken", ctx, "fakeClientID", "fakeClientSecret", "", "http://fakeServerDomain:8080/oauth/discord").Return("", errors.New("authorization code not found")).Once()

	token, _, errorURL, err := useCase.ValidateStateAndLogin(ctx, "", "valid_state")

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Equal(t, "http://fakeFrontendDomain:3000/?error=invalid_state", errorURL)

	mockStateStore.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "GetSetting")
}

// Test: Discord OAuth 错误（GetAccessToken 失败）
func TestDiscordLoginUseCase_ValidateStateAndLogin_DiscordOAuthError(t *testing.T) {
	_, _, useCase, mockDiscordOAuth, _, mockStateStore := setup()
	ctx := context.Background()

	mockStateStore.On("ValidateState", ctx, "valid_state").Return(nil).Once()
	mockDiscordOAuth.On("GetAccessToken", ctx, "fakeClientID", "fakeClientSecret", "invalid_code", "http://fakeServerDomain:8080/oauth/discord").Return("", errors.New("invalid authorization code")).Once()

	token, _, errorURL, err := useCase.ValidateStateAndLogin(ctx, "invalid_code", "valid_state")

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Equal(t, "http://fakeFrontendDomain:3000/?error=invalid_state", errorURL)

	mockStateStore.AssertExpectations(t)
	mockDiscordOAuth.AssertExpectations(t)
}

// Test: Discord GetGuildMemberData 错误
func TestDiscordLoginUseCase_ValidateStateAndLogin_GuildMemberDataError(t *testing.T) {
	_, _, useCase, mockDiscordOAuth, _, mockStateStore := setup()
	ctx := context.Background()

	mockStateStore.On("ValidateState", ctx, "valid_state").Return(nil).Once()
	mockDiscordOAuth.On("GetAccessToken", ctx, "fakeClientID", "fakeClientSecret", "valid_code", "http://fakeServerDomain:8080/oauth/discord").Return("access_token", nil).Once()
	mockDiscordOAuth.On("GetGuildMemberData", ctx, "access_token", "fakeGuildID").Return((*dto.DiscordGuildMemberDTO)(nil), errors.New("user not in guild")).Once()

	token, _, errorURL, err := useCase.ValidateStateAndLogin(ctx, "valid_code", "valid_state")

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Equal(t, "http://fakeFrontendDomain:3000/?error=invalid_state", errorURL)

	mockStateStore.AssertExpectations(t)
	mockDiscordOAuth.AssertExpectations(t)
}

// Test: HTTPS 配置下的 URL 生成
func TestDiscordLoginUseCase_InitiateLogin_HTTPS(t *testing.T) {
	mockRepo := new(mock_data.MockSystemSettingRepository)
	mockLogger := new(mock_data.MockLogger)
	mockDiscordOAuth := new(mock_data.MockDiscordOAuth)
	mockJWTGenerator := new(mock_data.MockJWTGenerator)
	mockStateStore := new(mock_data.MockStateStore)

	// HTTPS config
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
			Port:         443,
			Domain:       "asmr.pabo.live",
			HTTPS:        true, // HTTPS enabled
			EnableGinLog: true,
		},
		Frontend: struct {
			Domain string `mapstructure:"domain"`
			Port   int    `mapstructure:"port"`
		}{
			Domain: "asmr.pabo.live",
			Port:   443,
		},
		JWT: struct {
			SecretKey string `mapstructure:"secretKey"`
		}{
			SecretKey: "fakeJWTSecret",
		},
	}

	useCase := usecase.NewDiscordLoginUseCase(mockRepo, mockLogger, cfg, mockDiscordOAuth, mockJWTGenerator, mockStateStore)
	ctx := context.Background()

	mockStateStore.On("GenerateState", ctx).Return("https_test_state", nil).Once()

	authURL, err := useCase.InitiateLogin(ctx)

	assert.NoError(t, err)
	assert.Contains(t, authURL, "https://discord.com/api/oauth2/authorize")
	assert.Contains(t, authURL, "redirect_uri=https%3A%2F%2Fasmr.pabo.live%2Foauth%2Fdiscord")
	assert.Contains(t, authURL, "state=https_test_state")

	mockStateStore.AssertExpectations(t)
}

// Test: JWT 生成错误
func TestDiscordLoginUseCase_ValidateStateAndLogin_JWTGenerationError(t *testing.T) {
	mockRepo, _, useCase, mockDiscordOAuth, mockJWTGenerator, mockStateStore := setup()
	ctx := context.Background()

	testSetting := &system.Setting{
		EditorRoleId:        "editor123",
		StreamAccessRoleIds: []string{"user123", "user456"},
	}

	mockStateStore.On("ValidateState", ctx, "valid_state").Return(nil).Once()
	mockRepo.On("GetSetting").Return(testSetting, nil).Once()
	mockDiscordOAuth.On("GetAccessToken", ctx, "fakeClientID", "fakeClientSecret", "valid_code", "http://fakeServerDomain:8080/oauth/discord").Return("access_token", nil).Once()
	mockDiscordOAuth.On("GetGuildMemberData", ctx, "access_token", "fakeGuildID").Return(&dto.DiscordGuildMemberDTO{
		User:  dto.DiscordUserDTO{ID: "admin123"},
		Roles: []string{},
	}, nil).Once()
	mockJWTGenerator.On("GenerateDiscordToken", ctx, "admin123", mock.AnythingOfType("*dto.DiscordGuildMemberDTO"), role.Admin, "fakeJWTSecret").Return("", errors.New("jwt signing failed")).Once()

	token, _, errorURL, err := useCase.ValidateStateAndLogin(ctx, "valid_code", "valid_state")

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Equal(t, "http://fakeFrontendDomain:3000/?error=invalid_state", errorURL)

	mockStateStore.AssertExpectations(t)
	mockJWTGenerator.AssertExpectations(t)
}
