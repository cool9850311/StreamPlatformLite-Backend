package usecase

import (
	"Go-Service/src/main/application/dto/config"
	livestreamDto "Go-Service/src/main/application/dto/livestream"
	"Go-Service/src/main/application/usecase"
	"Go-Service/src/main/domain/entity/livestream"
	"Go-Service/src/test/usecase/mock_data"
	"context"
	"github.com/cool9850311/StreamPlatformLite-Core/pkg/role"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"Go-Service/src/main/domain/entity/chat"
	"Go-Service/src/main/domain/entity/errors"
	"Go-Service/src/main/infrastructure/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ================================================================================
// Test Setup
// ================================================================================

// Define a struct to hold all the mock objects
type LivestreamTestSetup struct {
	MockRepo             *mock_data.MockLivestreamRepository
	MockLogger           *mock_data.MockLogger
	MockViewerCountCache *mock_data.MockViewerCountCache
	MockChatCache        *mock_data.MockChatCache
	MockFileCache        *mock_data.MockFileCache
	MockFfmpegLibrary    *mock_data.MockFfmpegLibrary
	UseCase              *usecase.LivestreamUsecase
}

func setupLivestream() *LivestreamTestSetup {
	mockRepo := new(mock_data.MockLivestreamRepository)
	mockLogger := new(mock_data.MockLogger)
	mockStreamService := new(mock_data.MockLivestreamService)
	mockViewerCountCache := new(mock_data.MockViewerCountCache)
	mockChatCache := new(mock_data.MockChatCache)
	mockFileCache := new(mock_data.MockFileCache)
	mockFfmpegLibrary := new(mock_data.MockFfmpegLibrary)
	cfg := config.Config{
		Server: struct {
			Port         int    `mapstructure:"port"`
			Domain       string `mapstructure:"domain"`
			RTMPHost     string `mapstructure:"rtmp_host"`
			HTTPS        bool   `mapstructure:"https" default:"false"`
			EnableGinLog bool   `mapstructure:"enable_gin_log" default:"true"`
			LogLevel     string `mapstructure:"log_level" default:"INFO"`
		}{
			Domain:   "localhost",
			Port:     8080,
			HTTPS:    false,
			LogLevel: "INFO",
		},
	}
	useCase := usecase.NewLivestreamUsecase(mockRepo, mockLogger, cfg, mockStreamService, mockViewerCountCache, mockChatCache, mockFileCache, mockFfmpegLibrary)

	return &LivestreamTestSetup{
		MockRepo:             mockRepo,
		MockLogger:           mockLogger,
		MockViewerCountCache: mockViewerCountCache,
		MockChatCache:        mockChatCache,
		MockFileCache:        mockFileCache,
		MockFfmpegLibrary:    mockFfmpegLibrary,
		UseCase:              useCase,
	}
}

// ================================================================================
// API: GetLivestreamByID (5 tests)
// ================================================================================

// Role: Admin
func TestGetLivestreamByID_Admin_Success(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID: "livestream123",
	}

	setup.MockRepo.On("GetByID", "livestream123").Return(testLivestream, nil)

	result, err := setup.UseCase.GetLivestreamByID(ctx, "livestream123", role.Admin)
	assert.NotNil(t, result)
	assert.IsType(t, &livestreamDto.LivestreamGetByOwnerIDResponseDTO{}, result)
	assert.NoError(t, err)
	setup.MockRepo.AssertExpectations(t)
}

// Role: Editor (Unauthorized)
func TestGetLivestreamByID_Editor_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	result, err := setup.UseCase.GetLivestreamByID(ctx, "livestream123", role.Editor)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
	assert.Nil(t, result)
}

// Role: User (Unauthorized)
func TestGetLivestreamByID_User_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	result, err := setup.UseCase.GetLivestreamByID(ctx, "livestream123", role.User)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
	assert.Nil(t, result)
}

// Role: Guest (Unauthorized)
func TestGetLivestreamByID_Guest_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	result, err := setup.UseCase.GetLivestreamByID(ctx, "livestream123", role.Guest)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
	assert.Nil(t, result)
}

// Role: Anonymous (Unauthorized)
func TestGetLivestreamByID_Anonymous_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	result, err := setup.UseCase.GetLivestreamByID(ctx, "livestream123", role.Anonymous)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
	assert.Nil(t, result)
}

// ================================================================================
// API: GetLivestreamByOwnerID (5 tests)
// ================================================================================

// Role: Admin
func TestGetLivestreamByOwnerID_Admin_Success(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID: "livestream123",
	}

	setup.MockRepo.On("GetByOwnerID", "user123").Return(testLivestream, nil)

	result, err := setup.UseCase.GetLivestreamByOwnerID(ctx, "user123", role.Admin)

	assert.Equal(t, testLivestream.UUID, result.UUID)
	assert.NoError(t, err)
	setup.MockRepo.AssertExpectations(t)
}

// Role: Editor (Unauthorized)
func TestGetLivestreamByOwnerID_Editor_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	result, err := setup.UseCase.GetLivestreamByOwnerID(ctx, "owner123", role.Editor)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
	assert.Nil(t, result)
}

// Role: User (Unauthorized)
func TestGetLivestreamByOwnerID_User_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	result, err := setup.UseCase.GetLivestreamByOwnerID(ctx, "user123", role.User)

	assert.Nil(t, result)
	assert.Error(t, err)
}

// Role: Guest (Unauthorized)
func TestGetLivestreamByOwnerID_Guest_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	result, err := setup.UseCase.GetLivestreamByOwnerID(ctx, "owner123", role.Guest)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
	assert.Nil(t, result)
}

// Role: Anonymous (Unauthorized)
func TestGetLivestreamByOwnerID_Anonymous_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	result, err := setup.UseCase.GetLivestreamByOwnerID(ctx, "owner123", role.Anonymous)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
	assert.Nil(t, result)
}

// ================================================================================
// API: CreateLivestream (6 tests)
// ================================================================================

// Role: Admin - Success
func TestCreateLivestream_Admin_Success(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestreamDto.LivestreamCreateDTO{
		Name:        "Test Livestream",
		Title:       "Test Livestream",
		Information: "Test Livestream",
		Visibility:  "public",
	}
	setup.MockRepo.On("GetOne").Return(nil, errors.ErrNotFound)
	setup.MockRepo.On("Create", mock.AnythingOfType("*livestream.Livestream")).Return(nil)

	_, err := setup.UseCase.CreateLivestream(ctx, testLivestream, "user123", role.Admin)

	assert.NoError(t, err)
	setup.MockRepo.AssertExpectations(t)
}

// Role: Admin - Already Exists
func TestCreateLivestream_Admin_AlreadyExists(t *testing.T) {
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

// Role: User (Unauthorized)
func TestCreateLivestream_User_Unauthorized(t *testing.T) {
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

// Role: Editor (Unauthorized)
func TestCreateLivestream_Editor_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	livestreamData := &livestreamDto.LivestreamCreateDTO{
		Name:        "Test Stream",
		Visibility:  livestream.Public,
		Title:       "Test Title",
		Information: "Test Info",
	}

	result, err := setup.UseCase.CreateLivestream(ctx, livestreamData, "user123", role.Editor)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
	assert.Nil(t, result)
}

// Role: Guest (Unauthorized)
func TestCreateLivestream_Guest_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	livestreamData := &livestreamDto.LivestreamCreateDTO{
		Name:        "Test Stream",
		Visibility:  livestream.Public,
		Title:       "Test Title",
		Information: "Test Info",
	}

	result, err := setup.UseCase.CreateLivestream(ctx, livestreamData, "user123", role.Guest)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
	assert.Nil(t, result)
}

// Role: Anonymous (Unauthorized)
func TestCreateLivestream_Anonymous_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	livestreamData := &livestreamDto.LivestreamCreateDTO{
		Name:        "Test Stream",
		Visibility:  livestream.Public,
		Title:       "Test Title",
		Information: "Test Info",
	}

	result, err := setup.UseCase.CreateLivestream(ctx, livestreamData, "user123", role.Anonymous)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
	assert.Nil(t, result)
}

// ================================================================================
// API: UpdateLivestream (5 tests)
// ================================================================================

// Role: Admin
func TestUpdateLivestream_Admin_Success(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:  "livestream123",
		Title: "Test Livestream",
	}

	setup.MockRepo.On("Update", testLivestream).Return(nil)

	err := setup.UseCase.UpdateLivestream(ctx, testLivestream, role.Admin)

	assert.NoError(t, err)
	setup.MockRepo.AssertExpectations(t)
}

// Role: Editor (Unauthorized)
func TestUpdateLivestream_Editor_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:        "livestream123",
		Name:        "Updated Name",
		Visibility:  livestream.Public,
		Title:       "Updated Title",
		Information: "Updated Info",
	}

	err := setup.UseCase.UpdateLivestream(ctx, testLivestream, role.Editor)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
}

// Role: User (Unauthorized)
func TestUpdateLivestream_User_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:        "livestream123",
		Name:        "Updated Name",
		Visibility:  livestream.Public,
		Title:       "Updated Title",
		Information: "Updated Info",
	}

	err := setup.UseCase.UpdateLivestream(ctx, testLivestream, role.User)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
}

// Role: Guest (Unauthorized)
func TestUpdateLivestream_Guest_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:        "livestream123",
		Name:        "Updated Name",
		Visibility:  livestream.Public,
		Title:       "Updated Title",
		Information: "Updated Info",
	}

	err := setup.UseCase.UpdateLivestream(ctx, testLivestream, role.Guest)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
}

// Role: Anonymous (Unauthorized)
func TestUpdateLivestream_Anonymous_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:        "livestream123",
		Name:        "Updated Name",
		Visibility:  livestream.Public,
		Title:       "Updated Title",
		Information: "Updated Info",
	}

	err := setup.UseCase.UpdateLivestream(ctx, testLivestream, role.Anonymous)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
}

// ================================================================================
// API: DeleteLivestream (5 tests)
// ================================================================================

// Role: Admin
func TestDeleteLivestream_Admin_Success(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	setup.MockRepo.On("Delete", "livestream123").Return(nil)

	err := setup.UseCase.DeleteLivestream(ctx, "livestream123", role.Admin)

	assert.NoError(t, err)
	setup.MockRepo.AssertExpectations(t)
}

// Role: Editor (Unauthorized)
func TestDeleteLivestream_Editor_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	err := setup.UseCase.DeleteLivestream(ctx, "livestream123", role.Editor)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
}

// Role: User (Unauthorized)
func TestDeleteLivestream_User_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	err := setup.UseCase.DeleteLivestream(ctx, "livestream123", role.User)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
}

// Role: Guest (Unauthorized)
func TestDeleteLivestream_Guest_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	err := setup.UseCase.DeleteLivestream(ctx, "livestream123", role.Guest)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
}

// Role: Anonymous (Unauthorized)
func TestDeleteLivestream_Anonymous_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	err := setup.UseCase.DeleteLivestream(ctx, "livestream123", role.Anonymous)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
}

// ================================================================================
// API: GetOne (7 tests)
// Grouped by: Visibility -> Role
// ================================================================================

// Visibility: Public - Role: User
func TestGetOne_Public_User_Success(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:        "livestream123",
		Name:        "Test Livestream",
		Title:       "Test Title",
		Information: "Test Information",
		Visibility:  livestream.Public,
	}

	setup.MockRepo.On("GetOne").Return(testLivestream, nil)

	result, err := setup.UseCase.GetOne(ctx, role.User)

	expectedURL := "http://localhost:8080/livestream/livestream123/playlist.m3u8"
	assert.NoError(t, err)
	assert.Equal(t, testLivestream.UUID, result.UUID)
	assert.Equal(t, testLivestream.Name, result.Name)
	assert.Equal(t, testLivestream.Title, result.Title)
	assert.Equal(t, testLivestream.Information, result.Information)
	assert.Equal(t, expectedURL, result.StreamURL)
	setup.MockRepo.AssertExpectations(t)
}

// Visibility: Public - Role: Guest
func TestGetOne_Public_Guest_Success(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:        "livestream123",
		Name:        "Test Livestream",
		Title:       "Test Title",
		Information: "Test Information",
		Visibility:  livestream.Public,
	}

	setup.MockRepo.On("GetOne").Return(testLivestream, nil)

	result, err := setup.UseCase.GetOne(ctx, role.Guest)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, testLivestream.UUID, result.UUID)
	setup.MockRepo.AssertExpectations(t)
}

// Visibility: Public - Role: Anonymous
func TestGetOne_Public_Anonymous_Success(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:        "livestream123",
		Name:        "Test Livestream",
		Title:       "Test Title",
		Information: "Test Information",
		Visibility:  livestream.Public,
	}

	setup.MockRepo.On("GetOne").Return(testLivestream, nil)

	result, err := setup.UseCase.GetOne(ctx, role.Anonymous)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, testLivestream.UUID, result.UUID)
	assert.Equal(t, livestream.Public, result.Visibility)
	setup.MockRepo.AssertExpectations(t)
}

// Visibility: MemberOnly - Role: User
func TestGetOne_MemberOnly_User_Success(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:        "livestream123",
		Name:        "Test Livestream",
		Title:       "Test Title",
		Information: "Test Information",
		Visibility:  livestream.MemberOnly,
	}

	setup.MockRepo.On("GetOne").Return(testLivestream, nil)

	result, err := setup.UseCase.GetOne(ctx, role.User)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, testLivestream.UUID, result.UUID)
	setup.MockRepo.AssertExpectations(t)
}

// Visibility: MemberOnly - Role: Guest (Unauthorized)
func TestGetOne_MemberOnly_Guest_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:       "livestream123",
		Visibility: livestream.MemberOnly,
	}

	setup.MockRepo.On("GetOne").Return(testLivestream, nil)

	result, err := setup.UseCase.GetOne(ctx, role.Guest)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, errors.ErrUnauthorized, err)
	setup.MockRepo.AssertExpectations(t)
}

// Visibility: MemberOnly - Role: Anonymous (Unauthorized)
func TestGetOne_MemberOnly_Anonymous_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:       "livestream123",
		Visibility: livestream.MemberOnly,
	}

	setup.MockRepo.On("GetOne").Return(testLivestream, nil)

	result, err := setup.UseCase.GetOne(ctx, role.Anonymous)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, errors.ErrUnauthorized, err)
	setup.MockRepo.AssertExpectations(t)
}

// ================================================================================
// API: PingViewerCount (9 tests)
// Grouped by: Visibility -> Role
// ================================================================================

// Visibility: Public - Role: Admin
func TestPingViewerCount_Public_Admin_Success(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:       "livestream123",
		Visibility: livestream.Public,
	}

	setup.MockRepo.On("GetByID", "livestream123").Return(testLivestream, nil)
	setup.MockViewerCountCache.On("AddViewerCount", "livestream123", "user123").Return(nil)
	setup.MockViewerCountCache.On("GetViewerCount", "livestream123").Return(10, nil)

	viewerCount, err := setup.UseCase.PingViewerCount(ctx, role.Admin, "livestream123", "user123", "")

	assert.NoError(t, err)
	assert.Equal(t, 10, viewerCount)
	setup.MockRepo.AssertExpectations(t)
	setup.MockViewerCountCache.AssertExpectations(t)
}

// Visibility: Public - Role: User (Ignores AnonymousID)
func TestPingViewerCount_Public_User_IgnoresAnonymousID(t *testing.T) {
	setup := setupLivestream()
	defer setup.MockRepo.AssertExpectations(t)
	defer setup.MockViewerCountCache.AssertExpectations(t)

	livestreamUUID := "test-uuid-123"
	userID := "discord-user-123"
	anonymousID := "should-be-ignored" // Should not be used

	mockLivestream := &livestream.Livestream{
		UUID:       livestreamUUID,
		Visibility: livestream.Public,
	}

	// Mock expectations
	setup.MockRepo.On("GetByID", livestreamUUID).Return(mockLivestream, nil)
	setup.MockViewerCountCache.On("AddViewerCount", livestreamUUID, userID).Return(nil) // Uses userID, NOT anonymousID
	setup.MockViewerCountCache.On("GetViewerCount", livestreamUUID).Return(3, nil)

	// Execute
	count, err := setup.UseCase.PingViewerCount(
		context.Background(),
		role.User, // Logged-in user
		livestreamUUID,
		userID,
		anonymousID, // Should be ignored
	)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
	// Verify AddViewerCount was called with userID, not anonymousID
	setup.MockViewerCountCache.AssertCalled(t, "AddViewerCount", livestreamUUID, userID)
}

// Visibility: Public - Role: Anonymous (Valid ID)
func TestPingViewerCount_Public_Anonymous_WithValidID(t *testing.T) {
	setup := setupLivestream()
	defer setup.MockRepo.AssertExpectations(t)
	defer setup.MockViewerCountCache.AssertExpectations(t)

	livestreamUUID := "test-uuid-123"
	anonymousID := "550e8400-e29b-41d4-a716-446655440000" // UUID v4

	mockLivestream := &livestream.Livestream{
		UUID:       livestreamUUID,
		Visibility: livestream.Public,
	}

	// Mock expectations
	setup.MockRepo.On("GetByID", livestreamUUID).Return(mockLivestream, nil)
	setup.MockViewerCountCache.On("AddViewerCount", livestreamUUID, anonymousID).Return(nil)
	setup.MockViewerCountCache.On("GetViewerCount", livestreamUUID).Return(5, nil)

	// Execute
	count, err := setup.UseCase.PingViewerCount(
		context.Background(),
		role.Anonymous,
		livestreamUUID,
		"", // userID empty for anonymous
		anonymousID,
	)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, 5, count)
}

// Visibility: Public - Role: Anonymous (Without ID)
func TestPingViewerCount_Public_Anonymous_WithoutID(t *testing.T) {
	setup := setupLivestream()
	defer setup.MockRepo.AssertExpectations(t)

	livestreamUUID := "test-uuid-123"

	mockLivestream := &livestream.Livestream{
		UUID:       livestreamUUID,
		Visibility: livestream.Public,
	}

	// Mock expectations
	setup.MockRepo.On("GetByID", livestreamUUID).Return(mockLivestream, nil)
	// ViewerCountCache should NOT be called

	// Execute
	count, err := setup.UseCase.PingViewerCount(
		context.Background(),
		role.Anonymous,
		livestreamUUID,
		"", // userID empty
		"", // anonymousID empty - should fail
	)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, errors.ErrInvalidInput, err)
	assert.Equal(t, 0, count)
}

// Visibility: Public - Role: Anonymous (Empty String ID)
func TestPingViewerCount_Public_Anonymous_WithEmptyStringID(t *testing.T) {
	setup := setupLivestream()
	defer setup.MockRepo.AssertExpectations(t)

	livestreamUUID := "test-uuid-123"

	mockLivestream := &livestream.Livestream{
		UUID:       livestreamUUID,
		Visibility: livestream.Public,
	}

	setup.MockRepo.On("GetByID", livestreamUUID).Return(mockLivestream, nil)

	// Execute with empty string (not nil)
	count, err := setup.UseCase.PingViewerCount(
		context.Background(),
		role.Anonymous,
		livestreamUUID,
		"",
		"   ", // Whitespace-only should also fail
	)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, errors.ErrInvalidInput, err)
	assert.Equal(t, 0, count)
}

// Visibility: Public - Role: Anonymous (Same ID Updates Timestamp)
func TestPingViewerCount_Public_Anonymous_SameIDUpdatesTimestamp(t *testing.T) {
	setup := setupLivestream()
	defer setup.MockRepo.AssertExpectations(t)
	defer setup.MockViewerCountCache.AssertExpectations(t)

	livestreamUUID := "test-uuid-123"
	anonymousID := "550e8400-e29b-41d4-a716-446655440000"

	mockLivestream := &livestream.Livestream{
		UUID:       livestreamUUID,
		Visibility: livestream.Public,
	}

	// First ping
	setup.MockRepo.On("GetByID", livestreamUUID).Return(mockLivestream, nil).Once()
	setup.MockViewerCountCache.On("AddViewerCount", livestreamUUID, anonymousID).Return(nil).Once()
	setup.MockViewerCountCache.On("GetViewerCount", livestreamUUID).Return(1, nil).Once()

	count1, err1 := setup.UseCase.PingViewerCount(
		context.Background(),
		role.Anonymous,
		livestreamUUID,
		"",
		anonymousID,
	)

	// Second ping with same ID
	setup.MockRepo.On("GetByID", livestreamUUID).Return(mockLivestream, nil).Once()
	setup.MockViewerCountCache.On("AddViewerCount", livestreamUUID, anonymousID).Return(nil).Once()
	setup.MockViewerCountCache.On("GetViewerCount", livestreamUUID).Return(1, nil).Once() // Still 1

	count2, err2 := setup.UseCase.PingViewerCount(
		context.Background(),
		role.Anonymous,
		livestreamUUID,
		"",
		anonymousID,
	)

	// Assertions
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, 1, count1)
	assert.Equal(t, 1, count2) // Count unchanged
	// Verify AddViewerCount was called twice (updates timestamp)
	setup.MockViewerCountCache.AssertNumberOfCalls(t, "AddViewerCount", 2)
}

// Visibility: MemberOnly - Role: Guest (Unauthorized)
func TestPingViewerCount_MemberOnly_Guest_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:       "livestream123",
		Visibility: livestream.MemberOnly,
	}

	setup.MockRepo.On("GetByID", "livestream123").Return(testLivestream, nil)

	viewerCount, err := setup.UseCase.PingViewerCount(ctx, role.Guest, "livestream123", "user123", "")

	assert.Error(t, err)
	assert.Equal(t, 0, viewerCount)
	setup.MockRepo.AssertExpectations(t)
}

// Visibility: MemberOnly - Role: Anonymous (Unauthorized)
func TestPingViewerCount_MemberOnly_Anonymous_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	defer setup.MockRepo.AssertExpectations(t)

	livestreamUUID := "test-uuid-123"
	anonymousID := "550e8400-e29b-41d4-a716-446655440000"

	mockLivestream := &livestream.Livestream{
		UUID:       livestreamUUID,
		Visibility: livestream.MemberOnly, // Not public
	}

	// Mock expectations
	setup.MockRepo.On("GetByID", livestreamUUID).Return(mockLivestream, nil)
	// checkViewAccess should fail, so ViewerCountCache should NOT be called

	// Execute
	count, err := setup.UseCase.PingViewerCount(
		context.Background(),
		role.Anonymous,
		livestreamUUID,
		"",
		anonymousID,
	)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
	assert.Equal(t, 0, count)
}

// ================================================================================
// API: GetChat (4 tests)
// Grouped by: Visibility -> Role
// ================================================================================

// Visibility: Public - Role: Admin
func TestGetChat_Public_Admin_Success(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:       "livestream123",
		Visibility: livestream.Public,
	}

	testChats := []chat.Chat{
		{UserID: "user1", Message: "Hello", Avatar: "avatar1", Username: "username1", Role: 0},
		{UserID: "user2", Message: "Hi", Avatar: "avatar2", Username: "username2", Role: 3},
	}

	setup.MockRepo.On("GetByID", "livestream123").Return(testLivestream, nil)
	setup.MockChatCache.On("GetChat", "livestream123", "0", 10).Return(testChats, nil)

	chats, err := setup.UseCase.GetChat(ctx, role.Admin, "livestream123", "0")

	assert.NoError(t, err)
	assert.Equal(t, testChats, chats)
	setup.MockRepo.AssertExpectations(t)
	setup.MockChatCache.AssertExpectations(t)
}

// Visibility: Public - Role: Anonymous
func TestGetChat_Public_Anonymous_Success(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:       "livestream123",
		Visibility: livestream.Public,
	}

	testChats := []chat.Chat{
		{UserID: "user1", Message: "Hello", Role: role.User},
	}

	setup.MockRepo.On("GetByID", "livestream123").Return(testLivestream, nil)
	setup.MockChatCache.On("GetChat", "livestream123", "0", 10).Return(testChats, nil)

	chats, err := setup.UseCase.GetChat(ctx, role.Anonymous, "livestream123", "0")

	assert.NoError(t, err)
	assert.NotNil(t, chats)
	assert.Equal(t, testChats, chats)
	setup.MockRepo.AssertExpectations(t)
	setup.MockChatCache.AssertExpectations(t)
}

// Visibility: MemberOnly - Role: Guest (Unauthorized)
func TestGetChat_MemberOnly_Guest_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:       "livestream123",
		Visibility: livestream.MemberOnly,
	}

	setup.MockRepo.On("GetByID", "livestream123").Return(testLivestream, nil)

	chats, err := setup.UseCase.GetChat(ctx, role.Guest, "livestream123", "0")

	assert.Error(t, err)
	assert.Nil(t, chats)
	setup.MockRepo.AssertExpectations(t)
}

// ================================================================================
// API: AddChat (5 tests)
// Grouped by: Visibility -> Role
// ================================================================================

// Visibility: Public - Role: Admin
func TestAddChat_Public_Admin_Success(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:       "livestream123",
		Visibility: livestream.Public,
	}

	testChat := chat.Chat{UserID: "user123", Message: "Hello", Avatar: "avatar123", Username: "username123", Role: 0}

	setup.MockRepo.On("GetByID", "livestream123").Return(testLivestream, nil)
	setup.MockChatCache.On("AddChat", "livestream123", testChat).Return(nil)

	err := setup.UseCase.AddChat(ctx, "identityProvider", role.Admin, "livestream123", testChat)

	assert.NoError(t, err)
	setup.MockRepo.AssertExpectations(t)
	setup.MockChatCache.AssertExpectations(t)
}

// Visibility: Public - Role: Guest
func TestAddChat_Public_Guest_Success(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:       "livestream123",
		Visibility: livestream.Public,
	}

	testChat := chat.Chat{UserID: "guest123", Message: "Hello from guest", Role: role.Guest}

	setup.MockRepo.On("GetByID", "livestream123").Return(testLivestream, nil)
	setup.MockChatCache.On("AddChat", "livestream123", testChat).Return(nil)

	err := setup.UseCase.AddChat(ctx, "discord", role.Guest, "livestream123", testChat)

	assert.NoError(t, err)
	setup.MockRepo.AssertExpectations(t)
	setup.MockChatCache.AssertExpectations(t)
}

// Visibility: Public - Role: Anonymous (Unauthorized)
func TestAddChat_Public_Anonymous_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:       "livestream123",
		Visibility: livestream.Public,
	}

	testChat := chat.Chat{UserID: "user123", Message: "Hello", Role: role.Anonymous}

	setup.MockRepo.On("GetByID", "livestream123").Return(testLivestream, nil)

	err := setup.UseCase.AddChat(ctx, "test", role.Anonymous, "livestream123", testChat)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
	setup.MockRepo.AssertExpectations(t)
}

// Visibility: MemberOnly - Role: Guest (Unauthorized)
func TestAddChat_MemberOnly_Guest_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:       "livestream123",
		Visibility: livestream.MemberOnly,
	}

	testChat := chat.Chat{UserID: "guest123", Message: "Hello", Role: role.Guest}

	setup.MockRepo.On("GetByID", "livestream123").Return(testLivestream, nil)

	err := setup.UseCase.AddChat(ctx, "discord", role.Guest, "livestream123", testChat)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
	setup.MockRepo.AssertExpectations(t)
}

// ================================================================================
// API: DeleteChat (11 tests)
// Grouped by: Role
// ================================================================================

// Role: Admin - Can Delete Any
func TestDeleteChat_Admin_CanDeleteAny(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:       "livestream123",
		Visibility: livestream.Public,
	}

	// Admin can delete any chat message
	setup.MockRepo.On("GetByID", "livestream123").Return(testLivestream, nil)
	setup.MockChatCache.On("DeleteChat", "livestream123", "chat123").Return(nil)

	err := setup.UseCase.DeleteChat(ctx, role.Admin, "user456", "livestream123", "chat123")

	assert.NoError(t, err)
	setup.MockChatCache.AssertExpectations(t)
}

// Role: Editor - Can Delete User Message
func TestDeleteChat_Editor_CanDeleteUserMessage(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:       "livestream123",
		Visibility: livestream.Public,
	}

	// Mock getting a non-Admin chat message
	testChat := chat.Chat{
		ID:       "chat123",
		UserID:   "user123",
		Username: "Regular User",
		Role:     role.User,
		Message:  "Test message",
	}

	// Editor can delete non-Admin chat messages
	setup.MockRepo.On("GetByID", "livestream123").Return(testLivestream, nil)
	setup.MockChatCache.On("GetChatByID", "livestream123", "chat123").Return(&testChat, nil)
	setup.MockChatCache.On("DeleteChat", "livestream123", "chat123").Return(nil)

	err := setup.UseCase.DeleteChat(ctx, role.Editor, "user456", "livestream123", "chat123")

	assert.NoError(t, err)
	setup.MockRepo.AssertExpectations(t)
	setup.MockChatCache.AssertExpectations(t)
}

// Role: Editor - Cannot Delete Admin Message
func TestDeleteChat_Editor_CannotDeleteAdminMessage(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:       "livestream123",
		Visibility: livestream.Public,
	}

	// Mock: GetChatByID returns Admin's message
	adminChat := &chat.Chat{
		ID:       "admin-chat-001",
		UserID:   "admin-001",
		Username: "Admin User",
		Role:     role.Admin,
		Message:  "Admin message",
	}

	setup.MockRepo.On("GetByID", "livestream123").Return(testLivestream, nil)
	setup.MockChatCache.On("GetChatByID", "livestream123", "admin-chat-001").Return(adminChat, nil)

	// Editor tries to delete Admin's message
	err := setup.UseCase.DeleteChat(
		ctx,
		role.Editor,
		"editor-001",
		"livestream123",
		"admin-chat-001",
	)

	// Verify it returns Unauthorized
	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
	setup.MockRepo.AssertExpectations(t)
	setup.MockChatCache.AssertExpectations(t)
}

// Role: Editor - Cannot Delete Editor Message
func TestDeleteChat_Editor_CannotDeleteOtherEditorMessage(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:       "livestream123",
		Visibility: livestream.Public,
	}

	// Mock: GetChatByID returns another Editor's message
	editorChat := &chat.Chat{
		ID:       "editor-chat-001",
		UserID:   "editor-002",
		Username: "Another Editor",
		Role:     role.Editor,
		Message:  "Editor message",
	}

	setup.MockRepo.On("GetByID", "livestream123").Return(testLivestream, nil)
	setup.MockChatCache.On("GetChatByID", "livestream123", "editor-chat-001").Return(editorChat, nil)

	// Editor tries to delete another Editor's message
	err := setup.UseCase.DeleteChat(
		ctx,
		role.Editor,
		"editor-001",
		"livestream123",
		"editor-chat-001",
	)

	// Verify it returns Unauthorized
	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
	setup.MockRepo.AssertExpectations(t)
	setup.MockChatCache.AssertExpectations(t)
}

// Role: Editor - Can Delete Own Message
func TestDeleteChat_Editor_CanDeleteOwnMessage(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:       "livestream123",
		Visibility: livestream.Public,
	}

	// Mock: GetChatByID returns Editor's own message
	ownChat := &chat.Chat{
		ID:       "chat123",
		UserID:   "editor-001",
		Username: "Test Editor",
		Role:     role.Editor,
		Message:  "My own message",
	}

	setup.MockRepo.On("GetByID", "livestream123").Return(testLivestream, nil)
	setup.MockChatCache.On("GetChatByID", "livestream123", "chat123").Return(ownChat, nil)
	setup.MockChatCache.On("DeleteChat", "livestream123", "chat123").Return(nil)

	// Editor deletes their own message
	err := setup.UseCase.DeleteChat(
		ctx,
		role.Editor,
		"editor-001",
		"livestream123",
		"chat123",
	)

	// Verify it succeeds
	assert.NoError(t, err)
	setup.MockRepo.AssertExpectations(t)
	setup.MockChatCache.AssertExpectations(t)
}

// Role: User - Can Delete Own Message
func TestDeleteChat_User_CanDeleteOwnMessage(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:       "livestream123",
		Visibility: livestream.Public,
	}

	// Mock getting the chat to verify ownership
	testChat := chat.Chat{
		ID:       "chat123",
		UserID:   "user123",
		Avatar:   "avatar123",
		Username: "username123",
		Message:  "Test message",
		Role:     3,
	}
	setup.MockRepo.On("GetByID", "livestream123").Return(testLivestream, nil)
	setup.MockChatCache.On("GetChatByID", "livestream123", "chat123").Return(&testChat, nil)
	setup.MockChatCache.On("DeleteChat", "livestream123", "chat123").Return(nil)

	// User deletes their own message - should succeed
	err := setup.UseCase.DeleteChat(ctx, role.User, "user123", "livestream123", "chat123")

	assert.NoError(t, err)
	setup.MockRepo.AssertExpectations(t)
	setup.MockChatCache.AssertExpectations(t)
}

// Role: User - Cannot Delete Others Message
func TestDeleteChat_User_CannotDeleteOthersMessage(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:       "livestream123",
		Visibility: livestream.Public,
	}

	// Mock getting the chat to verify ownership
	testChat := chat.Chat{
		ID:       "chat123",
		UserID:   "user456", // Different user owns this message
		Avatar:   "avatar456",
		Username: "username456",
		Message:  "Test message",
		Role:     3,
	}
	setup.MockRepo.On("GetByID", "livestream123").Return(testLivestream, nil)
	setup.MockChatCache.On("GetChatByID", "livestream123", "chat123").Return(&testChat, nil)

	// User tries to delete someone else's message - should fail
	err := setup.UseCase.DeleteChat(ctx, role.User, "user123", "livestream123", "chat123")

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
	setup.MockRepo.AssertExpectations(t)
	setup.MockChatCache.AssertExpectations(t)
}

// Role: Guest - Can Delete Own Message
func TestDeleteChat_Guest_CanDeleteOwnMessage(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:       "livestream123",
		Visibility: livestream.Public,
	}

	testChat := chat.Chat{
		ID:      "chat123",
		UserID:  "guest123",
		Message: "Test message",
		Role:    role.Guest,
	}

	setup.MockRepo.On("GetByID", "livestream123").Return(testLivestream, nil)
	setup.MockChatCache.On("GetChatByID", "livestream123", "chat123").Return(&testChat, nil)
	setup.MockChatCache.On("DeleteChat", "livestream123", "chat123").Return(nil)

	err := setup.UseCase.DeleteChat(ctx, role.Guest, "guest123", "livestream123", "chat123")

	assert.NoError(t, err)
	setup.MockRepo.AssertExpectations(t)
	setup.MockChatCache.AssertExpectations(t)
}

// Role: Guest - Cannot Delete Others Message
func TestDeleteChat_Guest_CannotDeleteOthersMessage(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:       "livestream123",
		Visibility: livestream.Public,
	}

	testChat := chat.Chat{
		ID:      "chat123",
		UserID:  "otheruser",
		Message: "Test message",
		Role:    role.User,
	}

	setup.MockRepo.On("GetByID", "livestream123").Return(testLivestream, nil)
	setup.MockChatCache.On("GetChatByID", "livestream123", "chat123").Return(&testChat, nil)

	err := setup.UseCase.DeleteChat(ctx, role.Guest, "guest123", "livestream123", "chat123")

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
	setup.MockRepo.AssertExpectations(t)
	setup.MockChatCache.AssertExpectations(t)
}

// ================================================================================
// API: MuteUser (11 tests)
// Grouped by: Role
// ================================================================================

// Role: Admin - Can Mute Admin
func TestMuteUser_Admin_CanMuteOtherAdmin(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	// Mock chat with Admin role
	adminChat := &chat.Chat{
		ID:       "admin-chat-001",
		UserID:   "admin123",
		Username: "Admin User",
		Role:     role.Admin,
		Message:  "Admin message",
	}

	// Admin (role.Admin) mutes Admin (role.Admin) - should succeed
	setup.MockChatCache.On("GetChatByID", "livestream123", "admin-chat-001").Return(adminChat, nil)
	setup.MockRepo.On("MuteUser", "identityProvider", "livestream123", "admin123").Return(nil)

	err := setup.UseCase.MuteUser(ctx, "identityProvider", role.Admin, "admin-001", "livestream123", "admin-chat-001")

	assert.NoError(t, err)
	setup.MockChatCache.AssertExpectations(t)
	setup.MockRepo.AssertExpectations(t)
}

// Role: Admin - Can Mute Editor
func TestMuteUser_Admin_CanMuteEditor(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	// Mock chat with Editor role
	editorChat := &chat.Chat{
		ID:       "editor-chat-001",
		UserID:   "editor123",
		Username: "Editor User",
		Role:     role.Editor,
		Message:  "Editor message",
	}

	// Admin (role.Admin) mutes Editor (role.Editor) - should succeed
	setup.MockChatCache.On("GetChatByID", "livestream123", "editor-chat-001").Return(editorChat, nil)
	setup.MockRepo.On("MuteUser", "identityProvider", "livestream123", "editor123").Return(nil)

	err := setup.UseCase.MuteUser(ctx, "identityProvider", role.Admin, "admin-001", "livestream123", "editor-chat-001")

	assert.NoError(t, err)
	setup.MockChatCache.AssertExpectations(t)
	setup.MockRepo.AssertExpectations(t)
}

// Role: Admin - Cannot Mute Self
func TestMuteUser_Admin_CannotMuteSelf(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	// Mock chat with Admin role (trying to mute themselves)
	adminChat := &chat.Chat{
		ID:       "admin-chat-001",
		UserID:   "admin-001",
		Username: "Admin User",
		Role:     role.Admin,
		Message:  "Admin message",
	}

	// Admin tries to mute their own message
	setup.MockChatCache.On("GetChatByID", "livestream123", "admin-chat-001").Return(adminChat, nil)

	err := setup.UseCase.MuteUser(ctx, "identityProvider", role.Admin, "admin-001", "livestream123", "admin-chat-001")

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
	setup.MockChatCache.AssertExpectations(t)
	// Verify that MuteUser was NOT called on the repository
	setup.MockRepo.AssertNotCalled(t, "MuteUser")
}

// Role: Editor - Can Mute User
func TestMuteUser_Editor_CanMuteUser(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	// Mock chat with User role
	testChat := &chat.Chat{
		ID:       "chat123",
		UserID:   "user123",
		Username: "Regular User",
		Role:     role.User,
		Message:  "Test message",
	}

	// Editor (role.Editor) mutes User (role.User) - should succeed
	setup.MockChatCache.On("GetChatByID", "livestream123", "chat123").Return(testChat, nil)
	setup.MockRepo.On("MuteUser", "identityProvider", "livestream123", "user123").Return(nil)

	err := setup.UseCase.MuteUser(ctx, "identityProvider", role.Editor, "editor-001", "livestream123", "chat123")

	assert.NoError(t, err)
	setup.MockChatCache.AssertExpectations(t)
	setup.MockRepo.AssertExpectations(t)
}

// Role: Editor - Cannot Mute Admin
func TestMuteUser_Editor_CannotMuteAdmin(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	// Mock chat with Admin role
	adminChat := &chat.Chat{
		ID:       "admin-chat-001",
		UserID:   "admin123",
		Username: "Admin User",
		Role:     role.Admin,
		Message:  "Admin message",
	}

	// Editor (role.Editor) tries to mute Admin (role.Admin)
	// Should return errors.ErrUnauthorized
	// Should NOT call repository
	setup.MockChatCache.On("GetChatByID", "livestream123", "admin-chat-001").Return(adminChat, nil)

	err := setup.UseCase.MuteUser(ctx, "identityProvider", role.Editor, "editor-001", "livestream123", "admin-chat-001")

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
	setup.MockChatCache.AssertExpectations(t)
	// Verify that MuteUser was NOT called on the repository
	setup.MockRepo.AssertNotCalled(t, "MuteUser")
}

// Role: Editor - Cannot Mute Editor
func TestMuteUser_Editor_CannotMuteOtherEditor(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	// Mock chat with Editor role
	editorChat := &chat.Chat{
		ID:       "editor-chat-001",
		UserID:   "editor123",
		Username: "Editor User",
		Role:     role.Editor,
		Message:  "Editor message",
	}

	// Editor (role.Editor) tries to mute another Editor (role.Editor)
	// Should return errors.ErrUnauthorized
	// Should NOT call repository
	setup.MockChatCache.On("GetChatByID", "livestream123", "editor-chat-001").Return(editorChat, nil)

	err := setup.UseCase.MuteUser(ctx, "identityProvider", role.Editor, "editor-001", "livestream123", "editor-chat-001")

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
	setup.MockChatCache.AssertExpectations(t)
	// Verify that MuteUser was NOT called on the repository
	setup.MockRepo.AssertNotCalled(t, "MuteUser")
}

// Role: Editor - Cannot Mute Self
func TestMuteUser_Editor_CannotMuteSelf(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	// Mock chat with Editor role (trying to mute themselves)
	editorChat := &chat.Chat{
		ID:       "editor-chat-001",
		UserID:   "editor-001",
		Username: "Editor User",
		Role:     role.Editor,
		Message:  "Editor message",
	}

	// Editor tries to mute their own message
	setup.MockChatCache.On("GetChatByID", "livestream123", "editor-chat-001").Return(editorChat, nil)

	err := setup.UseCase.MuteUser(ctx, "identityProvider", role.Editor, "editor-001", "livestream123", "editor-chat-001")

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
	setup.MockChatCache.AssertExpectations(t)
	// Verify that MuteUser was NOT called on the repository
	setup.MockRepo.AssertNotCalled(t, "MuteUser")
}

// Role: User (Unauthorized)
func TestMuteUser_User_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	err := setup.UseCase.MuteUser(ctx, "identityProvider", role.User, "user-001", "livestream123", "chat123")

	assert.Error(t, err)
}

// ================================================================================
// API: GetFile (5 tests)
// Grouped by: Visibility -> Role
// ================================================================================

// Visibility: Public - Role: Anonymous
func TestGetFile_Public_Anonymous_Success(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:       "83636040-7f54-49f2-ae40-9a1213614729",
		Visibility: livestream.Public,
	}

	testFileData := []byte("test file content")

	setup.MockRepo.On("GetOne").Return(testLivestream, nil)
	setup.MockFileCache.On("LoadCache", "/test/root/hls/83636040-7f54-49f2-ae40-9a1213614729/playlist.m3u8").Return(testFileData, true)

	file, err := setup.UseCase.GetFile(
		ctx,
		"/test/root",                           // rootPath (trusted)
		"83636040-7f54-49f2-ae40-9a1213614729", // uuid (external)
		"playlist.m3u8",                        // filename (external)
		role.Anonymous,
	)

	assert.NoError(t, err)
	assert.Equal(t, testFileData, file)
	setup.MockRepo.AssertExpectations(t)
	setup.MockFileCache.AssertExpectations(t)
}

// Visibility: MemberOnly - Role: Guest (Unauthorized)
func TestGetFile_MemberOnly_Guest_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:       "83636040-7f54-49f2-ae40-9a1213614729",
		Visibility: livestream.MemberOnly,
	}

	setup.MockRepo.On("GetOne").Return(testLivestream, nil)

	file, err := setup.UseCase.GetFile(
		ctx,
		"/test/root",
		"83636040-7f54-49f2-ae40-9a1213614729",
		"playlist.m3u8",
		role.Guest,
	)

	assert.Error(t, err)
	assert.Nil(t, file)
	assert.Equal(t, errors.ErrUnauthorized, err)
	setup.MockRepo.AssertExpectations(t)
}

// Role: Editor (Unauthorized)
func TestGetRecord_Editor_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	result, err := setup.UseCase.GetRecord(ctx, "/test/root", "livestream123", role.Editor)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
	assert.Empty(t, result)
}

// Role: User (Unauthorized)
func TestGetRecord_User_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	result, err := setup.UseCase.GetRecord(ctx, "/test/root", "livestream123", role.User)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
	assert.Empty(t, result)
}

// Not Found Test
func TestGetFile_NotFound(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	// Test for record.m3u8 (blocked in usecase)
	file, err := setup.UseCase.GetFile(
		ctx,
		"/test/root",
		"83636040-7f54-49f2-ae40-9a1213614729",
		"record.m3u8",
		role.User,
	)
	assert.Nil(t, file)
	assert.Equal(t, errors.ErrNotFound, err)

	// Test for *.mp4 (should fail at validation)
	file, err = setup.UseCase.GetFile(
		ctx,
		"/test/root",
		"83636040-7f54-49f2-ae40-9a1213614729",
		"output.mp4",
		role.User,
	)
	assert.Nil(t, file)
	assert.Equal(t, errors.ErrInvalidInput, err)
}

// ================================================================================
// API: GetRecord (6 tests)
// Grouped by: Role
// ================================================================================

// Role: Admin - Success
func TestGetRecord_Admin_Success(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	validUUID := "83636040-7f54-49f2-ae40-9a1213614729"
	rootPath := "/test/root"
	filePath := "/test/root/hls/" + validUUID + "/*.mp4"

	// The fixed name is probed first; this test covers the legacy glob branch,
	// so the probe must miss.
	setup.MockFileCache.On("Exists", recordFixedPath(validUUID)).Return(false, nil)
	setup.MockFileCache.On("GetSingleFileName", filePath).Return("output.mp4", nil)
	// The display name is now resolved through the repository; an empty Title
	// keeps the on-disk basename, which is what this test is about.
	setup.MockRepo.On("GetByID", validUUID).Return(&livestream.Livestream{UUID: validUUID}, nil)

	result, err := setup.UseCase.GetRecord(ctx, rootPath, validUUID, role.Admin)

	assert.NoError(t, err)
	assert.Equal(t, "output.mp4", result.Path)
	setup.MockFileCache.AssertExpectations(t)
}

// Role: Admin - Not Mp4 (this test is no longer relevant as GetRecord always uses *.mp4)
// Removed - GetRecord now always constructs path with *.mp4

// Role: Guest (Unauthorized)
func TestGetRecord_Guest_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()
	result, err := setup.UseCase.GetRecord(ctx, "/test/root", "livestream123", role.Guest)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
	assert.Empty(t, result)
}

// Role: Anonymous (Unauthorized)
func TestGetRecord_Anonymous_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	result, err := setup.UseCase.GetRecord(ctx, "/test/root", "livestream123", role.Anonymous)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
	assert.Empty(t, result)
}

// ================================================================================
// API: GetDeleteChatIDs (4 tests)
// Grouped by: Visibility -> Role
// ================================================================================

// Visibility: Public - Role: Anonymous
func TestGetDeleteChatIDs_Public_Anonymous_Success(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:       "test-uuid",
		Visibility: livestream.Public,
	}

	deletedIDs := []string{"chat1", "chat2"}

	setup.MockRepo.On("GetByID", "test-uuid").Return(testLivestream, nil)
	setup.MockChatCache.On("GetDeleteChatIDs", "test-uuid").Return(deletedIDs, nil)

	result, err := setup.UseCase.GetDeleteChatIDs(ctx, role.Anonymous, "test-uuid")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, deletedIDs, result)
	setup.MockRepo.AssertExpectations(t)
	setup.MockChatCache.AssertExpectations(t)
}

// Visibility: MemberOnly - Role: Anonymous (Unauthorized)
func TestGetDeleteChatIDs_MemberOnly_Anonymous_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:       "test-uuid",
		Visibility: livestream.MemberOnly,
	}

	setup.MockRepo.On("GetByID", "test-uuid").Return(testLivestream, nil)

	result, err := setup.UseCase.GetDeleteChatIDs(ctx, role.Anonymous, "test-uuid")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, errors.ErrUnauthorized, err)
	setup.MockRepo.AssertExpectations(t)
}

// Visibility: Public - Role: Guest
func TestGetDeleteChatIDs_Public_Guest_Success(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:       "test-uuid",
		Visibility: livestream.Public,
	}

	deletedIDs := []string{"chat1", "chat2"}

	setup.MockRepo.On("GetByID", "test-uuid").Return(testLivestream, nil)
	setup.MockChatCache.On("GetDeleteChatIDs", "test-uuid").Return(deletedIDs, nil)

	result, err := setup.UseCase.GetDeleteChatIDs(ctx, role.Guest, "test-uuid")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, deletedIDs, result)
	setup.MockRepo.AssertExpectations(t)
	setup.MockChatCache.AssertExpectations(t)
}

// Visibility: MemberOnly - Role: Guest (Unauthorized)
func TestGetDeleteChatIDs_MemberOnly_Guest_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:       "test-uuid",
		Visibility: livestream.MemberOnly,
	}

	setup.MockRepo.On("GetByID", "test-uuid").Return(testLivestream, nil)

	result, err := setup.UseCase.GetDeleteChatIDs(ctx, role.Guest, "test-uuid")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, errors.ErrUnauthorized, err)
	setup.MockRepo.AssertExpectations(t)
}

// ================================================================================
// Path Validation Tests for GetFile (Table-Driven)
// ================================================================================

func TestGetFile_PathValidation(t *testing.T) {
	validUUID := "83636040-7f54-49f2-ae40-9a1213614729"
	rootPath := "/test/root"

	tests := []struct {
		name           string
		uuid           string
		filename       string
		expectError    error
		needMock       bool
		mockLivestream bool
	}{
		{
			name:           "Valid M3U8",
			uuid:           validUUID,
			filename:       "playlist.m3u8",
			expectError:    nil,
			needMock:       true,
			mockLivestream: true,
		},
		{
			name:           "Valid TS",
			uuid:           validUUID,
			filename:       "segment0.ts",
			expectError:    nil,
			needMock:       true,
			mockLivestream: true,
		},
		{
			name:        "Invalid UUID",
			uuid:        "invalid-uuid",
			filename:    "playlist.m3u8",
			expectError: errors.ErrInvalidInput,
		},
		{
			name:        "Path Traversal",
			uuid:        validUUID,
			filename:    "../../../etc/passwd",
			expectError: errors.ErrInvalidInput,
		},
		{
			name:        "Filename With Slash",
			uuid:        validUUID,
			filename:    "subdir/file.m3u8",
			expectError: errors.ErrInvalidInput,
		},
		{
			name:        "Mp4 Extension",
			uuid:        validUUID,
			filename:    "video.mp4",
			expectError: errors.ErrInvalidInput,
		},
		{
			name:        "Php Extension",
			uuid:        validUUID,
			filename:    "malicious.php",
			expectError: errors.ErrInvalidInput,
		},
		{
			name:        "Absolute Path",
			uuid:        validUUID,
			filename:    "/etc/passwd",
			expectError: errors.ErrInvalidInput,
		},
		{
			name:        "Empty Filename",
			uuid:        validUUID,
			filename:    "",
			expectError: errors.ErrInvalidInput,
		},
		{
			name:        "Record M3u8",
			uuid:        validUUID,
			filename:    "record.m3u8",
			expectError: errors.ErrNotFound,
		},
		{
			name:        "Backslash Separator",
			uuid:        validUUID,
			filename:    "..\\..\\etc\\passwd",
			expectError: errors.ErrInvalidInput,
		},
		{
			name:        "Double Slash",
			uuid:        validUUID,
			filename:    "//playlist.m3u8",
			expectError: errors.ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup := setupLivestream()
			ctx := context.Background()

			if tt.mockLivestream {
				testLivestream := &livestream.Livestream{
					UUID:       validUUID,
					Visibility: livestream.Public,
				}
				setup.MockRepo.On("GetOne").Return(testLivestream, nil)
			}

			if tt.needMock {
				testFileData := []byte("test file content")
				filePath := filepath.Join(rootPath, "hls", tt.uuid, tt.filename)
				setup.MockFileCache.On("LoadCache", filePath).Return(testFileData, true)
			}

			file, err := setup.UseCase.GetFile(ctx, rootPath, tt.uuid, tt.filename, role.User)

			if tt.expectError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectError, err)
				assert.Nil(t, file)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, file)
			}

			if tt.mockLivestream || tt.needMock {
				setup.MockRepo.AssertExpectations(t)
				setup.MockFileCache.AssertExpectations(t)
			}
		})
	}
}

// ================================================================================
// Path Validation Tests for GetRecord (Table-Driven)
// ================================================================================

func TestGetRecord_PathValidation(t *testing.T) {
	validUUID := "83636040-7f54-49f2-ae40-9a1213614729"
	rootPath := "/test/root"

	tests := []struct {
		name        string
		uuid        string
		expectError error
		needMock    bool
	}{
		{
			name:        "Valid Mp4",
			uuid:        validUUID,
			expectError: nil,
			needMock:    true,
		},
		{
			name:        "Invalid UUID",
			uuid:        "invalid-uuid",
			expectError: errors.ErrInvalidInput,
		},
		{
			name:        "Empty UUID",
			uuid:        "",
			expectError: errors.ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup := setupLivestream()
			ctx := context.Background()

			if tt.needMock {
				filePath := filepath.Join(rootPath, "hls", tt.uuid, "*.mp4")
				// Fixed-name probe misses so the legacy glob is used.
				setup.MockFileCache.On("Exists", recordFixedPath(tt.uuid)).Return(false, nil)
				setup.MockFileCache.On("GetSingleFileName", filePath).Return("output.mp4", nil)
				setup.MockRepo.On("GetByID", tt.uuid).Return(&livestream.Livestream{UUID: tt.uuid}, nil)
			}

			result, err := setup.UseCase.GetRecord(ctx, rootPath, tt.uuid, role.Admin)

			if tt.expectError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectError, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "output.mp4", result.Path)
			}

			if tt.needMock {
				setup.MockFileCache.AssertExpectations(t)
			}
		})
	}
}

// ================================================================================
// API: GetRecordStatus (12 tests) + GetRecord read-only guard (1 test)
// Grouped by: State machine -> Task table -> Role/Input validation
// ================================================================================

const (
	recordTestRootPath  = "/test/root"
	recordTestUUID      = "83636040-7f54-49f2-ae40-9a1213614729"
	recordTestOtherUUID = "9c2f1d3e-4a5b-4c6d-8e7f-0a1b2c3d4e5f"
)

// recordMp4Glob returns the glob the usecase uses to look for the converted mp4.
func recordMp4Glob(livestreamUUID string) string {
	return filepath.Join(recordTestRootPath, "hls", livestreamUUID, "*.mp4")
}

// recordFixedPath returns the fixed-name recording path the usecase probes first
// (read precedence step 1: util.RecordingFileName, then the "*.mp4" glob).
func recordFixedPath(livestreamUUID string) string {
	return filepath.Join(recordTestRootPath, "hls", livestreamUUID, util.RecordingFileName)
}

// recordPlaylistPath returns the HLS record playlist path the usecase probes.
func recordPlaylistPath(livestreamUUID string) string {
	return filepath.Join(recordTestRootPath, "hls", livestreamUUID, "record.m3u8")
}

// waitForFfmpegCalls polls the ffmpeg mock call counter until it reaches want,
// so that conversions started in a background goroutine can be observed without
// relying on a bare sleep. It fails the test when the bound is exceeded.
func waitForFfmpegCalls(t *testing.T, setup *LivestreamTestSetup, want int, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		if got := setup.MockFfmpegLibrary.CallCount(); got >= want {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out after %s waiting for ffmpeg call count to reach %d, got %d",
				timeout, want, setup.MockFfmpegLibrary.CallCount())
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// waitForConvertTaskRelease waits until the conversion task registered for
// livestreamUUID has been removed from the task table. Release is not directly
// observable from outside the usecase, so it is detected by the only externally
// visible consequence: a further GetRecordStatus call is able to start a brand
// new conversion, pushing the ffmpeg call counter past baseline.
//
// The caller is expected to have made the retry conversion blocking (SetDelay)
// so that at most one extra conversion can be started while polling.
func waitForConvertTaskRelease(t *testing.T, setup *LivestreamTestSetup, livestreamUUID string, baseline int, timeout time.Duration) {
	t.Helper()
	ctx := context.Background()
	deadline := time.Now().Add(timeout)
	for {
		if setup.MockFfmpegLibrary.CallCount() > baseline {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out after %s waiting for the convert task of %s to be released (ffmpeg calls stuck at %d)",
				timeout, livestreamUUID, setup.MockFfmpegLibrary.CallCount())
		}
		status, err := setup.UseCase.GetRecordStatus(ctx, recordTestRootPath, livestreamUUID, role.Admin)
		if err != nil {
			t.Fatalf("unexpected error while polling GetRecordStatus: %v", err)
		}
		if status.State != usecase.RecordStateConverting {
			t.Fatalf("expected state %q while polling, got %q", usecase.RecordStateConverting, status.State)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// State: ready
func TestGetRecordStatus_Admin_Ready(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	mp4Path := filepath.Join(recordTestRootPath, "hls", recordTestUUID, "output.mp4")
	setup.MockFileCache.On("Exists", recordFixedPath(recordTestUUID)).Return(false, nil)
	setup.MockFileCache.On("GetSingleFileName", recordMp4Glob(recordTestUUID)).Return(mp4Path, nil)
	setup.MockFileCache.On("Stat", mp4Path).Return(int64(4096), time.Time{}, nil)
	// Empty Title -> the reported filename is the on-disk basename, as before.
	setup.MockRepo.On("GetByID", recordTestUUID).Return(&livestream.Livestream{UUID: recordTestUUID}, nil)

	status, err := setup.UseCase.GetRecordStatus(ctx, recordTestRootPath, recordTestUUID, role.Admin)

	assert.NoError(t, err)
	assert.Equal(t, usecase.RecordStateReady, status.State)
	assert.Equal(t, "output.mp4", status.Filename)
	assert.Equal(t, int64(4096), status.SizeBytes)
	assert.Equal(t, 0, setup.MockFfmpegLibrary.CallCount(), "ffmpeg must not be called when the mp4 already exists")
	setup.MockFileCache.AssertExpectations(t)
}

// State: unavailable -> converting
func TestGetRecordStatus_Admin_TriggersConvert_WhenM3u8Exists(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	setup.MockFileCache.On("Exists", recordFixedPath(recordTestUUID)).Return(false, nil)
	setup.MockFileCache.On("GetSingleFileName", recordMp4Glob(recordTestUUID)).Return("", errors.ErrNotFound)
	setup.MockFileCache.On("Exists", recordPlaylistPath(recordTestUUID)).Return(true, nil)
	setup.MockRepo.On("GetByID", recordTestUUID).Return(&livestream.Livestream{
		UUID:  recordTestUUID,
		Title: "Test Record",
	}, nil)

	status, err := setup.UseCase.GetRecordStatus(ctx, recordTestRootPath, recordTestUUID, role.Admin)

	assert.NoError(t, err)
	assert.Equal(t, usecase.RecordStateConverting, status.State)
	assert.Equal(t, int64(0), status.ElapsedSeconds)

	waitForFfmpegCalls(t, setup, 1, 3*time.Second)
	assert.Equal(t, 1, setup.MockFfmpegLibrary.CallCount())
	setup.MockFileCache.AssertExpectations(t)
}

// State: unavailable
func TestGetRecordStatus_Admin_Unavailable_NoM3u8(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	setup.MockFileCache.On("Exists", recordFixedPath(recordTestUUID)).Return(false, nil)
	setup.MockFileCache.On("GetSingleFileName", recordMp4Glob(recordTestUUID)).Return("", errors.ErrNotFound)
	setup.MockFileCache.On("Exists", recordPlaylistPath(recordTestUUID)).Return(false, nil)

	status, err := setup.UseCase.GetRecordStatus(ctx, recordTestRootPath, recordTestUUID, role.Admin)

	assert.NoError(t, err)
	assert.Equal(t, usecase.RecordStateUnavailable, status.State)
	assert.Equal(t, 0, setup.MockFfmpegLibrary.CallCount(), "ffmpeg must not be called without a record playlist")
	setup.MockFileCache.AssertExpectations(t)
}

// Task table: same uuid is converted only once
func TestGetRecordStatus_SameUUID_OnlyOneConvertTask(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	setup.MockFfmpegLibrary.SetDelay(500 * time.Millisecond)
	setup.MockFileCache.On("Exists", recordFixedPath(recordTestUUID)).Return(false, nil)
	setup.MockFileCache.On("GetSingleFileName", recordMp4Glob(recordTestUUID)).Return("", errors.ErrNotFound)
	setup.MockFileCache.On("Exists", recordPlaylistPath(recordTestUUID)).Return(true, nil)
	setup.MockRepo.On("GetByID", recordTestUUID).Return(&livestream.Livestream{
		UUID:  recordTestUUID,
		Title: "Test Record",
	}, nil)

	for i := 0; i < 5; i++ {
		status, err := setup.UseCase.GetRecordStatus(ctx, recordTestRootPath, recordTestUUID, role.Admin)
		assert.NoError(t, err)
		assert.Equal(t, usecase.RecordStateConverting, status.State, "call %d should report converting", i+1)
	}

	waitForFfmpegCalls(t, setup, 1, 3*time.Second)
	assert.Equal(t, 1, setup.MockFfmpegLibrary.CallCount(), "the same uuid must only start one conversion")
	setup.MockFileCache.AssertExpectations(t)
}

// Task table: different uuids convert in parallel (regression test for the old global mutex)
func TestGetRecordStatus_DifferentUUIDs_BothConvert(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	setup.MockFfmpegLibrary.SetDelay(500 * time.Millisecond)
	for _, uuid := range []string{recordTestUUID, recordTestOtherUUID} {
		setup.MockFileCache.On("Exists", recordFixedPath(uuid)).Return(false, nil)
		setup.MockFileCache.On("GetSingleFileName", recordMp4Glob(uuid)).Return("", errors.ErrNotFound)
		setup.MockFileCache.On("Exists", recordPlaylistPath(uuid)).Return(true, nil)
		setup.MockRepo.On("GetByID", uuid).Return(&livestream.Livestream{
			UUID:  uuid,
			Title: "Test Record " + uuid,
		}, nil)
	}

	firstStatus, err := setup.UseCase.GetRecordStatus(ctx, recordTestRootPath, recordTestUUID, role.Admin)
	assert.NoError(t, err)
	assert.Equal(t, usecase.RecordStateConverting, firstStatus.State)

	secondStatus, err := setup.UseCase.GetRecordStatus(ctx, recordTestRootPath, recordTestOtherUUID, role.Admin)
	assert.NoError(t, err)
	assert.Equal(t, usecase.RecordStateConverting, secondStatus.State, "a second uuid must not be blocked by the first conversion")

	waitForFfmpegCalls(t, setup, 2, 3*time.Second)
	assert.Equal(t, 2, setup.MockFfmpegLibrary.CallCount(), "each uuid must get its own conversion")
	setup.MockFileCache.AssertExpectations(t)
}

// State: converting -> ready
func TestGetRecordStatus_ConvertFinished_ThenReady(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	glob := recordMp4Glob(recordTestUUID)
	setup.MockFileCache.On("Exists", recordFixedPath(recordTestUUID)).Return(false, nil)
	missCall := setup.MockFileCache.On("GetSingleFileName", glob).Return("", errors.ErrNotFound)
	setup.MockFileCache.On("Exists", recordPlaylistPath(recordTestUUID)).Return(true, nil)
	// The convert path no longer reads the livestream; this stub now only feeds
	// the display name of the "ready" phase below, where an empty Title keeps
	// the on-disk basename this test asserts on.
	setup.MockRepo.On("GetByID", recordTestUUID).Return(&livestream.Livestream{
		UUID: recordTestUUID,
	}, nil)

	status, err := setup.UseCase.GetRecordStatus(ctx, recordTestRootPath, recordTestUUID, role.Admin)
	assert.NoError(t, err)
	assert.Equal(t, usecase.RecordStateConverting, status.State)

	// The first conversion runs to completion, then its task entry is released.
	waitForFfmpegCalls(t, setup, 1, 3*time.Second)
	setup.MockFfmpegLibrary.SetDelay(500 * time.Millisecond)
	waitForConvertTaskRelease(t, setup, recordTestUUID, 1, 3*time.Second)

	// Re-point the file cache so the mp4 glob now hits.
	mp4Path := filepath.Join(recordTestRootPath, "hls", recordTestUUID, "output.mp4")
	missCall.Unset()
	setup.MockFileCache.On("GetSingleFileName", glob).Return(mp4Path, nil)
	setup.MockFileCache.On("Stat", mp4Path).Return(int64(2048), time.Time{}, nil)

	status, err = setup.UseCase.GetRecordStatus(ctx, recordTestRootPath, recordTestUUID, role.Admin)
	assert.NoError(t, err)
	assert.Equal(t, usecase.RecordStateReady, status.State)
	assert.Equal(t, "output.mp4", status.Filename)
	assert.Equal(t, int64(2048), status.SizeBytes)
}

// Role: Editor (Unauthorized)
func TestGetRecordStatus_Editor_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	status, err := setup.UseCase.GetRecordStatus(ctx, recordTestRootPath, recordTestUUID, role.Editor)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
	assert.Equal(t, usecase.RecordStatus{}, status)
}

// Role: User (Unauthorized)
func TestGetRecordStatus_User_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	status, err := setup.UseCase.GetRecordStatus(ctx, recordTestRootPath, recordTestUUID, role.User)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
	assert.Equal(t, usecase.RecordStatus{}, status)
}

// Role: Guest (Unauthorized)
func TestGetRecordStatus_Guest_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	status, err := setup.UseCase.GetRecordStatus(ctx, recordTestRootPath, recordTestUUID, role.Guest)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
	assert.Equal(t, usecase.RecordStatus{}, status)
}

// Role: Anonymous (Unauthorized)
func TestGetRecordStatus_Anonymous_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	status, err := setup.UseCase.GetRecordStatus(ctx, recordTestRootPath, recordTestUUID, role.Anonymous)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
	assert.Equal(t, usecase.RecordStatus{}, status)
}

// Input validation: invalid uuid
func TestGetRecordStatus_InvalidUUID(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	status, err := setup.UseCase.GetRecordStatus(ctx, recordTestRootPath, "not-a-uuid", role.Admin)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrInvalidInput, err)
	assert.Equal(t, usecase.RecordStatus{}, status)
	assert.Equal(t, 0, setup.MockFfmpegLibrary.CallCount())
}

// Task table: a failed conversion releases its task entry
func TestGetRecordStatus_ConvertFails_TaskReleased(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	setup.MockFfmpegLibrary.SetError(errors.ErrInternal)
	setup.MockFileCache.On("Exists", recordFixedPath(recordTestUUID)).Return(false, nil)
	setup.MockFileCache.On("GetSingleFileName", recordMp4Glob(recordTestUUID)).Return("", errors.ErrNotFound)
	setup.MockFileCache.On("Exists", recordPlaylistPath(recordTestUUID)).Return(true, nil)
	setup.MockRepo.On("GetByID", recordTestUUID).Return(&livestream.Livestream{
		UUID:  recordTestUUID,
		Title: "Test Record",
	}, nil)

	status, err := setup.UseCase.GetRecordStatus(ctx, recordTestRootPath, recordTestUUID, role.Admin)
	assert.NoError(t, err)
	assert.Equal(t, usecase.RecordStateConverting, status.State)

	waitForFfmpegCalls(t, setup, 1, 3*time.Second)

	// Make the retry conversion blocking so exactly one new task can be started.
	setup.MockFfmpegLibrary.SetDelay(500 * time.Millisecond)
	waitForConvertTaskRelease(t, setup, recordTestUUID, 1, 3*time.Second)

	assert.Equal(t, 2, setup.MockFfmpegLibrary.CallCount(), "a failed conversion must release its task so a retry can start")
	setup.MockFileCache.AssertExpectations(t)
}

// A4 guard: GetRecord is read-only
func TestGetRecord_NoLongerTriggersConvert(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	setup.MockFileCache.On("Exists", recordFixedPath(recordTestUUID)).Return(false, nil)
	setup.MockFileCache.On("GetSingleFileName", recordMp4Glob(recordTestUUID)).Return("", errors.ErrNotFound)

	result, err := setup.UseCase.GetRecord(ctx, recordTestRootPath, recordTestUUID, role.Admin)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrNotFound, err)
	assert.Empty(t, result)
	assert.Equal(t, 0, setup.MockFfmpegLibrary.CallCount(), "GetRecord must never trigger a conversion")
	setup.MockFileCache.AssertExpectations(t)
	setup.MockRepo.AssertNotCalled(t, "GetByID")
}

// ================================================================================
// Read precedence and display name (plan tasks C-1 .. C-4)
// ================================================================================

// C-1: the fixed name wins over a legacy "*.mp4" file.
func TestGetRecordStatus_PrefersFixedName(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	fixedPath := recordFixedPath(recordTestUUID)
	setup.MockFileCache.On("Exists", fixedPath).Return(true, nil)
	setup.MockFileCache.On("Stat", fixedPath).Return(int64(1024), time.Time{}, nil)
	// Empty Title, so the reported filename is the basename of whichever file was
	// resolved - that is what makes the assertion below able to tell the two
	// candidates apart.
	setup.MockRepo.On("GetByID", recordTestUUID).Return(&livestream.Livestream{UUID: recordTestUUID}, nil)

	// Deliberately NO expectation for GetSingleFileName: a legacy file would be
	// resolvable on disk, so if the usecase consulted the glob at all the mock
	// would fail the test with an unexpected call.

	status, err := setup.UseCase.GetRecordStatus(ctx, recordTestRootPath, recordTestUUID, role.Admin)

	assert.NoError(t, err)
	assert.Equal(t, usecase.RecordStateReady, status.State)
	assert.Equal(t, util.RecordingFileName, status.Filename, "the fixed name must win over any legacy *.mp4")
	assert.Equal(t, int64(1024), status.SizeBytes, "the size must come from the fixed-name file")
	setup.MockFileCache.AssertNotCalled(t, "GetSingleFileName", mock.Anything)
	setup.MockFileCache.AssertExpectations(t)
}

// C-2: with no fixed-name file, a legacy "*.mp4" is still reported as ready.
func TestGetRecordStatus_LegacyFileFallback(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	legacyPath := filepath.Join(recordTestRootPath, "hls", recordTestUUID, "legacy record.mp4")
	setup.MockFileCache.On("Exists", recordFixedPath(recordTestUUID)).Return(false, nil)
	setup.MockFileCache.On("GetSingleFileName", recordMp4Glob(recordTestUUID)).Return(legacyPath, nil)
	setup.MockFileCache.On("Stat", legacyPath).Return(int64(7777), time.Time{}, nil)
	setup.MockRepo.On("GetByID", recordTestUUID).Return(&livestream.Livestream{UUID: recordTestUUID}, nil)

	status, err := setup.UseCase.GetRecordStatus(ctx, recordTestRootPath, recordTestUUID, role.Admin)

	assert.NoError(t, err)
	assert.Equal(t, usecase.RecordStateReady, status.State, "a legacy recording must stay downloadable")
	assert.Equal(t, int64(7777), status.SizeBytes, "the size must be stat'ed on the legacy file")
	assert.Equal(t, filepath.Base(legacyPath), status.Filename)
	assert.Equal(t, 0, setup.MockFfmpegLibrary.CallCount(), "a legacy recording must not trigger a conversion")
	setup.MockFileCache.AssertExpectations(t)
}

// C-3: the reported filename comes from the Title, not from the on-disk name.
func TestGetRecordStatus_FilenameFromTitle(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	// A title the sanitiser has to change: the path separator is dropped.
	const title = "Live/Show 2026"

	mp4Path := filepath.Join(recordTestRootPath, "hls", recordTestUUID, "output.mp4")
	setup.MockFileCache.On("Exists", recordFixedPath(recordTestUUID)).Return(false, nil)
	setup.MockFileCache.On("GetSingleFileName", recordMp4Glob(recordTestUUID)).Return(mp4Path, nil)
	setup.MockFileCache.On("Stat", mp4Path).Return(int64(512), time.Time{}, nil)
	setup.MockRepo.On("GetByID", recordTestUUID).Return(&livestream.Livestream{
		UUID:  recordTestUUID,
		Title: title,
	}, nil)

	status, err := setup.UseCase.GetRecordStatus(ctx, recordTestRootPath, recordTestUUID, role.Admin)

	assert.NoError(t, err)
	assert.Equal(t, usecase.RecordStateReady, status.State)
	assert.Equal(t, util.SanitizeDownloadFilename(title), status.Filename)
	// Pinned literal: this is the one case that documents the contract, so a
	// silent change in how separators are handled has to break the test.
	assert.Equal(t, "LiveShow 2026.mp4", status.Filename)
	assert.NotEqual(t, filepath.Base(mp4Path), status.Filename, "the on-disk name must stay invisible to the client")
	setup.MockFileCache.AssertExpectations(t)
}

// C-4: a blank Title degrades to the on-disk basename.
func TestGetRecordStatus_FilenameFallsBackToDiskName_WhenTitleEmpty(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	mp4Path := filepath.Join(recordTestRootPath, "hls", recordTestUUID, "legacy-2026.mp4")
	setup.MockFileCache.On("Exists", recordFixedPath(recordTestUUID)).Return(false, nil)
	setup.MockFileCache.On("GetSingleFileName", recordMp4Glob(recordTestUUID)).Return(mp4Path, nil)
	setup.MockFileCache.On("Stat", mp4Path).Return(int64(64), time.Time{}, nil)
	setup.MockRepo.On("GetByID", recordTestUUID).Return(&livestream.Livestream{
		UUID:  recordTestUUID,
		Title: "   \t  ",
	}, nil)

	status, err := setup.UseCase.GetRecordStatus(ctx, recordTestRootPath, recordTestUUID, role.Admin)

	assert.NoError(t, err)
	assert.Equal(t, usecase.RecordStateReady, status.State)
	assert.Equal(t, filepath.Base(mp4Path), status.Filename, "a blank title must fall back to the on-disk basename")
	setup.MockFileCache.AssertExpectations(t)
}

// ================================================================================
// Title validation at the write entrypoints (plan tasks C-5, C-6)
// ================================================================================

// invalidTitleCases are the inputs util.ValidateLivestreamTitle must reject.
func invalidTitleCases() []struct {
	name  string
	title string
} {
	return []struct {
		name  string
		title string
	}{
		{name: "Control Character", title: "evil\r\ntitle"},
		{name: "Too Long", title: strings.Repeat("a", util.MaxTitleRunes+1)},
	}
}

// C-5
func TestCreateLivestream_RejectsInvalidTitle(t *testing.T) {
	for _, tt := range invalidTitleCases() {
		t.Run(tt.name, func(t *testing.T) {
			setup := setupLivestream()
			ctx := context.Background()

			result, err := setup.UseCase.CreateLivestream(ctx, &livestreamDto.LivestreamCreateDTO{
				Name:        "Test Stream",
				Visibility:  livestream.Public,
				Title:       tt.title,
				Information: "Test Info",
			}, "user123", role.Admin)

			assert.Error(t, err)
			assert.Equal(t, errors.ErrInvalidInput, err)
			assert.Nil(t, result)
			setup.MockRepo.AssertNotCalled(t, "Create", mock.Anything)
			setup.MockRepo.AssertNotCalled(t, "GetOne")
		})
	}
}

// C-6
func TestUpdateLivestream_RejectsInvalidTitle(t *testing.T) {
	for _, tt := range invalidTitleCases() {
		t.Run(tt.name, func(t *testing.T) {
			setup := setupLivestream()
			ctx := context.Background()

			err := setup.UseCase.UpdateLivestream(ctx, &livestream.Livestream{
				UUID:  recordTestUUID,
				Title: tt.title,
			}, role.Admin)

			assert.Error(t, err)
			assert.Equal(t, errors.ErrInvalidInput, err)
			setup.MockRepo.AssertNotCalled(t, "Update", mock.Anything)
		})
	}
}
