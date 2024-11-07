package usecase

import (
	"Go-Service/src/main/application/dto/config"
	livestreamDto "Go-Service/src/main/application/dto/livestream"
	"Go-Service/src/main/application/usecase"
	"Go-Service/src/main/domain/entity/livestream"
	"Go-Service/src/main/domain/entity/role"
	"Go-Service/src/test/usecase/mock_data"
	"context"
	"testing"

	"Go-Service/src/main/domain/entity/chat"
	"Go-Service/src/main/domain/entity/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Define a struct to hold all the mock objects
type LivestreamTestSetup struct {
	MockRepo            *mock_data.MockLivestreamRepository
	MockLogger          *mock_data.MockLogger
	MockViewerCountCache *mock_data.MockViewerCountCache
	MockChatCache       *mock_data.MockChatCache
	MockFileCache       *mock_data.MockFileCache
	UseCase             usecase.LivestreamUsecase
}

func setupLivestream() *LivestreamTestSetup {
	mockRepo := new(mock_data.MockLivestreamRepository)
	mockLogger := new(mock_data.MockLogger)
	mockStreamService := new(mock_data.MockLivestreamService)
	mockViewerCountCache := new(mock_data.MockViewerCountCache)
	mockChatCache := new(mock_data.MockChatCache)
	mockFileCache := new(mock_data.MockFileCache)
	cfg := config.Config{
		Server: struct {
			Port         int    `mapstructure:"port"`
			Domain       string `mapstructure:"domain"`
			HTTPS        bool   `mapstructure:"https" default:"false"`
			EnableGinLog bool   `mapstructure:"enable_gin_log" default:"true"`
		}{
			Domain: "localhost",
			Port:   8080,
			HTTPS:  false,
		},
	}
	useCase := usecase.NewLivestreamUsecase(mockRepo, mockLogger, cfg, mockStreamService, mockViewerCountCache, mockChatCache, mockFileCache)

	return &LivestreamTestSetup{
		MockRepo:            mockRepo,
		MockLogger:          mockLogger,
		MockViewerCountCache: mockViewerCountCache,
		MockChatCache:       mockChatCache,
		MockFileCache:       mockFileCache,
		UseCase:             *useCase, // Return the pointer directly
	}
}

func TestLivestreamUsecase_GetLivestreamByID_AdminUser(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID: "livestream123",
		// other fields...
	}

	setup.MockRepo.On("GetByID", "livestream123").Return(testLivestream, nil)

	result, err := setup.UseCase.GetLivestreamByID(ctx, "livestream123", role.Admin)

	assert.Equal(t, testLivestream, result)
	assert.NoError(t, err)
	setup.MockRepo.AssertExpectations(t)
}

func TestLivestreamUsecase_GetLivestreamByID_UnauthorizedUser(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	result, err := setup.UseCase.GetLivestreamByID(ctx, "livestream123", role.User)

	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestLivestreamUsecase_GetLivestreamByOwnerID_AdminUser(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID: "livestream123",
		// other fields...
	}

	setup.MockRepo.On("GetByOwnerID", "user123").Return(testLivestream, nil)

	result, err := setup.UseCase.GetLivestreamByOwnerID(ctx, "user123", role.Admin)

	assert.Equal(t, testLivestream.UUID, result.UUID)
	assert.NoError(t, err)
	setup.MockRepo.AssertExpectations(t)
}

func TestLivestreamUsecase_GetLivestreamByOwnerID_UnauthorizedUser(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	result, err := setup.UseCase.GetLivestreamByOwnerID(ctx, "user123", role.User)

	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestLivestreamUsecase_CreateLivestream_AdminUser(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestreamDto.LivestreamCreateDTO{
		Name:        "Test Livestream",
		Title:       "Test Livestream",
		Information: "Test Livestream",
		Visibility:  "public",
	}
	setup.MockRepo.On("GetOne").Return(nil, errors.ErrNotFound) // it is actually document not found
	setup.MockRepo.On("Create", mock.AnythingOfType("*livestream.Livestream")).Return(nil)

	_, err := setup.UseCase.CreateLivestream(ctx, testLivestream, "user123", role.Admin)

	assert.NoError(t, err)
	setup.MockRepo.AssertExpectations(t)
}

func TestLivestreamUsecase_CreateLivestream_UnauthorizedUser(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()
	testLivestream := &livestreamDto.LivestreamCreateDTO{
		Name:        "Test Livestream",
		Title:       "Test Livestream",
		Information: "Test Livestream",
		Visibility:  "public",
	}

	_, err := setup.UseCase.CreateLivestream(ctx, testLivestream, "user123", role.User)

	assert.Error(t, err)
}

func TestLivestreamUsecase_CreateLivestream_AlreadyExists(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestreamDto.LivestreamCreateDTO{
		Name:        "Test Livestream",
		Title:       "Test Livestream",
		Information: "Test Livestream",
		Visibility:  "public",
	}

	// Mock GetOne to return a livestream, indicating it already exists
	setup.MockRepo.On("GetOne").Return(&livestream.Livestream{}, nil)

	_, err := setup.UseCase.CreateLivestream(ctx, testLivestream, "user123", role.Admin)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrExists, err)
	setup.MockRepo.AssertExpectations(t)
}

func TestLivestreamUsecase_UpdateLivestream_AdminUser(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:  "livestream123",
		Title: "Test Livestream",
		// other fields...
	}

	setup.MockRepo.On("Update", testLivestream).Return(nil)

	err := setup.UseCase.UpdateLivestream(ctx, testLivestream, role.Admin)

	assert.NoError(t, err)
	setup.MockRepo.AssertExpectations(t)
}

func TestLivestreamUsecase_UpdateLivestream_UnauthorizedUser(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID: "livestream123",
		// other fields...
	}

	err := setup.UseCase.UpdateLivestream(ctx, testLivestream, role.User)

	assert.Error(t, err)
}

func TestLivestreamUsecase_DeleteLivestream_AdminUser(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	setup.MockRepo.On("Delete", "livestream123").Return(nil)

	err := setup.UseCase.DeleteLivestream(ctx, "livestream123", role.Admin)

	assert.NoError(t, err)
	setup.MockRepo.AssertExpectations(t)
}

func TestLivestreamUsecase_DeleteLivestream_UnauthorizedUser(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	err := setup.UseCase.DeleteLivestream(ctx, "livestream123", role.User)

	assert.Error(t, err)
}

func TestLivestreamUsecase_GetOne_UserRole(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:           "livestream123",
		Name:           "Test Livestream",
		Title:          "Test Title",
		Information:    "Test Information",
		OutputPathUUID: "output123",
	}

	setup.MockRepo.On("GetOne").Return(testLivestream, nil)

	result, err := setup.UseCase.GetOne(ctx, role.User)

	expectedURL := "http://localhost:8080/livestream/output123/playlist.m3u8"
	assert.NoError(t, err)
	assert.Equal(t, testLivestream.UUID, result.UUID)
	assert.Equal(t, testLivestream.Name, result.Name)
	assert.Equal(t, testLivestream.Title, result.Title)
	assert.Equal(t, testLivestream.Information, result.Information)
	assert.Equal(t, expectedURL, result.StreamURL)
	setup.MockRepo.AssertExpectations(t)
}

func TestLivestreamUsecase_GetOne_UnauthorizedRole(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	result, err := setup.UseCase.GetOne(ctx, role.Guest)

	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestLivestreamUsecase_PingViewerCount_AdminUser(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	setup.MockViewerCountCache.On("AddViewerCount", "livestream123", "user123").Return(nil)
	setup.MockViewerCountCache.On("GetViewerCount", "livestream123").Return(10, nil)

	viewerCount, err := setup.UseCase.PingViewerCount(ctx, role.Admin, "livestream123", "user123")

	assert.NoError(t, err)
	assert.Equal(t, 10, viewerCount)
	setup.MockViewerCountCache.AssertExpectations(t)
}

func TestLivestreamUsecase_PingViewerCount_UnauthorizedUser(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	viewerCount, err := setup.UseCase.PingViewerCount(ctx, role.Guest, "livestream123", "user123")

	assert.Error(t, err)
	assert.Equal(t, 0, viewerCount)
}

func TestLivestreamUsecase_GetChat_AdminUser(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testChats := []chat.Chat{
		{UserID: "user1", Message: "Hello", Avatar: "avatar1", Username: "username1"},
		{UserID: "user2", Message: "Hi", Avatar: "avatar2", Username: "username2"},
	}

	setup.MockChatCache.On("GetChat", "livestream123", "0", 10).Return(testChats, nil)

	chats, err := setup.UseCase.GetChat(ctx, role.Admin, "livestream123", "0")

	assert.NoError(t, err)
	assert.Equal(t, testChats, chats)
	setup.MockChatCache.AssertExpectations(t)
}

func TestLivestreamUsecase_GetChat_UnauthorizedUser(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	chats, err := setup.UseCase.GetChat(ctx, role.Guest, "livestream123", "0")

	assert.Error(t, err)
	assert.Nil(t, chats)
}

func TestLivestreamUsecase_AddChat_AdminUser(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testChat := chat.Chat{UserID: "user123", Message: "Hello", Avatar: "avatar123", Username: "username123"}

	setup.MockRepo.On("GetByID", "livestream123").Return(&livestream.Livestream{}, nil)
	setup.MockChatCache.On("AddChat", "livestream123", testChat).Return(nil)

	err := setup.UseCase.AddChat(ctx, role.Admin, "livestream123", testChat)

	assert.NoError(t, err)
	setup.MockChatCache.AssertExpectations(t)
}

func TestLivestreamUsecase_AddChat_UnauthorizedUser(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testChat := chat.Chat{UserID: "user123", Message: "Hello"}

	err := setup.UseCase.AddChat(ctx, role.Guest, "livestream123", testChat)

	assert.Error(t, err)
}

func TestLivestreamUsecase_DeleteChat_EditorUser(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	setup.MockChatCache.On("DeleteChat", "livestream123", "chat123").Return(nil)

	err := setup.UseCase.DeleteChat(ctx, role.Editor, "livestream123", "chat123")

	assert.NoError(t, err)
	setup.MockChatCache.AssertExpectations(t)
}

func TestLivestreamUsecase_DeleteChat_UnauthorizedUser(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	err := setup.UseCase.DeleteChat(ctx, role.User, "livestream123", "chat123")

	assert.Error(t, err)
}

func TestLivestreamUsecase_MuteUser_EditorUser(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	setup.MockRepo.On("MuteUser", "livestream123", "user123").Return(nil)

	err := setup.UseCase.MuteUser(ctx, role.Editor, "livestream123", "user123")

	assert.NoError(t, err)
	setup.MockRepo.AssertExpectations(t)
}

func TestLivestreamUsecase_MuteUser_UnauthorizedUser(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	err := setup.UseCase.MuteUser(ctx, role.User, "livestream123", "user123")

	assert.Error(t, err)
}
