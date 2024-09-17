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

	"Go-Service/src/main/domain/entity/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupLivestream() (*mock_data.MockLivestreamRepository, *mock_data.MockLogger, usecase.LivestreamUsecase) {
	mockRepo := new(mock_data.MockLivestreamRepository)
	mockLogger := new(mock_data.MockLogger)
	mockStreamService := new(mock_data.MockLivestreamService)
	cfg := config.Config{
		Server: struct {
			Port   int    `mapstructure:"port"`
			Domain string `mapstructure:"domain"`
			HTTPS  bool   `mapstructure:"https" default:"false"`
		}{
			Domain: "localhost",
			Port:   8080,
			HTTPS:  false,
		},
	}
	useCase := usecase.NewLivestreamUsecase(mockRepo, mockLogger, cfg, mockStreamService)
	return mockRepo, mockLogger, *useCase // Dereference the pointer
}

func TestLivestreamUsecase_GetLivestreamByID_AdminUser(t *testing.T) {
	mockRepo, _, useCase := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID: "livestream123",
		// other fields...
	}

	mockRepo.On("GetByID", "livestream123").Return(testLivestream, nil)

	result, err := useCase.GetLivestreamByID(ctx, "livestream123", role.Admin)

	assert.Equal(t, testLivestream, result)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestLivestreamUsecase_GetLivestreamByID_UnauthorizedUser(t *testing.T) {
	_, _, useCase := setupLivestream()
	ctx := context.Background()

	result, err := useCase.GetLivestreamByID(ctx, "livestream123", role.User)

	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestLivestreamUsecase_GetLivestreamByOwnerID_AdminUser(t *testing.T) {
	mockRepo, _, useCase := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID: "livestream123",
		// other fields...
	}

	mockRepo.On("GetByOwnerID", "user123").Return(testLivestream, nil)

	result, err := useCase.GetLivestreamByOwnerID(ctx, "user123", role.Admin)

	assert.Equal(t, testLivestream.UUID, result.UUID)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestLivestreamUsecase_GetLivestreamByOwnerID_UnauthorizedUser(t *testing.T) {
	_, _, useCase := setupLivestream()
	ctx := context.Background()

	result, err := useCase.GetLivestreamByOwnerID(ctx, "user123", role.User)

	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestLivestreamUsecase_CreateLivestream_AdminUser(t *testing.T) {
	mockRepo, _, useCase := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestreamDto.LivestreamCreateDTO{
		Name:        "Test Livestream",
		Title:       "Test Livestream",
		Information: "Test Livestream",
		Visibility:  "public",
	}
	mockRepo.On("GetOne").Return(nil, errors.ErrNotFound) // it is actually document not found
	mockRepo.On("Create", mock.AnythingOfType("*livestream.Livestream")).Return(nil)

	_, err := useCase.CreateLivestream(ctx, testLivestream, "user123", role.Admin)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestLivestreamUsecase_CreateLivestream_UnauthorizedUser(t *testing.T) {
	_, _, useCase := setupLivestream()
	ctx := context.Background()
	testLivestream := &livestreamDto.LivestreamCreateDTO{
		Name:        "Test Livestream",
		Title:       "Test Livestream",
		Information: "Test Livestream",
		Visibility:  "public",
	}

	_, err := useCase.CreateLivestream(ctx, testLivestream, "user123", role.User)

	assert.Error(t, err)
}

func TestLivestreamUsecase_CreateLivestream_AlreadyExists(t *testing.T) {
	mockRepo, _, useCase := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestreamDto.LivestreamCreateDTO{
		Name:        "Test Livestream",
		Title:       "Test Livestream",
		Information: "Test Livestream",
		Visibility:  "public",
	}

	// Mock GetOne to return a livestream, indicating it already exists
	mockRepo.On("GetOne").Return(&livestream.Livestream{}, nil)

	_, err := useCase.CreateLivestream(ctx, testLivestream, "user123", role.Admin)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrExists, err)
	mockRepo.AssertExpectations(t)
}

func TestLivestreamUsecase_UpdateLivestream_AdminUser(t *testing.T) {
	mockRepo, _, useCase := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:  "livestream123",
		Title: "Test Livestream",
		// other fields...
	}

	mockRepo.On("Update", testLivestream).Return(nil)

	err := useCase.UpdateLivestream(ctx, testLivestream, role.Admin)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestLivestreamUsecase_UpdateLivestream_UnauthorizedUser(t *testing.T) {
	_, _, useCase := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID: "livestream123",
		// other fields...
	}

	err := useCase.UpdateLivestream(ctx, testLivestream, role.User)

	assert.Error(t, err)
}

func TestLivestreamUsecase_DeleteLivestream_AdminUser(t *testing.T) {
	mockRepo, _, useCase := setupLivestream()
	ctx := context.Background()

	mockRepo.On("Delete", "livestream123").Return(nil)

	err := useCase.DeleteLivestream(ctx, "livestream123", role.Admin)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestLivestreamUsecase_DeleteLivestream_UnauthorizedUser(t *testing.T) {
	_, _, useCase := setupLivestream()
	ctx := context.Background()

	err := useCase.DeleteLivestream(ctx, "livestream123", role.User)

	assert.Error(t, err)
}

func TestLivestreamUsecase_GetOne_UserRole(t *testing.T) {
	mockRepo, _, useCase := setupLivestream()
	ctx := context.Background()

	testLivestream := &livestream.Livestream{
		UUID:           "livestream123",
		Name:           "Test Livestream",
		Title:          "Test Title",
		Information:    "Test Information",
		OutputPathUUID: "output123",
	}

	mockRepo.On("GetOne").Return(testLivestream, nil)

	result, err := useCase.GetOne(ctx, role.User)

	expectedURL := "http://localhost:8080/livestream/output123/playlist.m3u8"
	assert.NoError(t, err)
	assert.Equal(t, testLivestream.UUID, result.UUID)
	assert.Equal(t, testLivestream.Name, result.Name)
	assert.Equal(t, testLivestream.Title, result.Title)
	assert.Equal(t, testLivestream.Information, result.Information)
	assert.Equal(t, expectedURL, result.StreamURL)
	mockRepo.AssertExpectations(t)
}

func TestLivestreamUsecase_GetOne_UnauthorizedRole(t *testing.T) {
	_, _, useCase := setupLivestream()
	ctx := context.Background()

	result, err := useCase.GetOne(ctx, role.Guest)

	assert.Nil(t, result)
	assert.Error(t, err)
}
