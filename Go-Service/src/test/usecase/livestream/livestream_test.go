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
		UUID:       "livestream123",
		Visibility: livestream.Public,
	}

	testFileData := []byte("test file content")

	setup.MockRepo.On("GetOne").Return(testLivestream, nil)
	setup.MockFileCache.On("LoadCache", "playlist.m3u8").Return(testFileData, true)

	file, err := setup.UseCase.GetFile(ctx, "playlist.m3u8", role.Anonymous)

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
		UUID:       "livestream123",
		Visibility: livestream.MemberOnly,
	}

	setup.MockRepo.On("GetOne").Return(testLivestream, nil)

	file, err := setup.UseCase.GetFile(ctx, "playlist.m3u8", role.Guest)

	assert.Error(t, err)
	assert.Nil(t, file)
	assert.Equal(t, errors.ErrUnauthorized, err)
	setup.MockRepo.AssertExpectations(t)
}

// Role: Editor (Unauthorized)
func TestGetRecord_Editor_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	result, err := setup.UseCase.GetRecord(ctx, "livestream123", "*.mp4", role.Editor)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
	assert.Empty(t, result)
}

// Role: User (Unauthorized)
func TestGetRecord_User_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	result, err := setup.UseCase.GetRecord(ctx, "livestream123", "*.mp4", role.User)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
	assert.Empty(t, result)
}


// Not Found Test
func TestGetFile_NotFound(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:       "livestream123",
		Visibility: livestream.Public,
	}

	setup.MockRepo.On("GetOne").Return(testLivestream, nil)

	// Test for record.m3u8
	file, err := setup.UseCase.GetFile(ctx, "record.m3u8", role.User)
	assert.Nil(t, file)
	assert.Equal(t, errors.ErrNotFound, err)

	// Test for *.mp4
	file, err = setup.UseCase.GetFile(ctx, "output.mp4", role.User)
	assert.Nil(t, file)
	assert.Equal(t, errors.ErrNotFound, err)

	setup.MockRepo.AssertExpectations(t)
}

// ================================================================================
// API: GetRecord (6 tests)
// Grouped by: Role
// ================================================================================

// Role: Admin - Success
func TestGetRecord_Admin_Success(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()
	setup.MockFileCache.On("GetSingleFileName", "*.mp4").Return("output.mp4", nil)

	result, err := setup.UseCase.GetRecord(ctx, "livestream123", "*.mp4", role.Admin)

	assert.NoError(t, err)
	assert.Equal(t, "output.mp4", result)
	setup.MockFileCache.AssertExpectations(t)
}

// Role: Admin - Not Mp4
func TestGetRecord_Admin_NotMp4(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	result, err := setup.UseCase.GetRecord(ctx, "livestream123", "*.m3u8", role.Admin)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrNotFound, err)
	assert.Empty(t, result)
}

// Role: Guest (Unauthorized)
func TestGetRecord_Guest_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()
	result, err := setup.UseCase.GetRecord(ctx, "livestream123", "*.mp4", role.Guest)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnauthorized, err)
	assert.Empty(t, result)
}

// Role: Anonymous (Unauthorized)
func TestGetRecord_Anonymous_Unauthorized(t *testing.T) {
	setup := setupLivestream()
	ctx := context.Background()

	result, err := setup.UseCase.GetRecord(ctx, "livestream123", "*.mp4", role.Anonymous)

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
