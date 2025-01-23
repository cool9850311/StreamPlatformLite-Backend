package origin_account

import (
	"context"
	"errors"
	"testing"

	"Go-Service/src/main/application/usecase"
	"Go-Service/src/main/domain/entity/account"
	"Go-Service/src/main/domain/entity/role"
	"Go-Service/src/test/usecase/mock_data"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"Go-Service/src/main/application/dto/config"
	innerErrors "Go-Service/src/main/domain/entity/errors"
)

// Define a struct to hold all the mock objects
type OriginAccountTestSetup struct {
	MockRepo   *mock_data.MockAccountRepository
	MockLogger *mock_data.MockLogger
	MockBcrypt *mock_data.MockBcryptGenerator
	MockJWTGenerator *mock_data.MockJWTGenerator
	UseCase    usecase.OriginAccountUseCase
}

func setupOriginAccount() *OriginAccountTestSetup {
	mockRepo := new(mock_data.MockAccountRepository)
	mockLogger := new(mock_data.MockLogger)
	mockBcrypt := new(mock_data.MockBcryptGenerator)
	mockJWTGenerator := new(mock_data.MockJWTGenerator)
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
	useCase := usecase.NewOriginAccountUseCase(mockRepo, mockLogger, mockBcrypt, cfg, mockJWTGenerator)

	return &OriginAccountTestSetup{
		MockRepo:   mockRepo,
		MockLogger: mockLogger,
		MockBcrypt: mockBcrypt,
		MockJWTGenerator: mockJWTGenerator,
		UseCase:    *useCase,
	}
}

func TestOriginAccountUseCase_Login_Success(t *testing.T) {
	setup := setupOriginAccount()
	ctx := context.Background()

	mockAccount := &account.Account{ID: "testuser", Username: "testuser", Password: "hashedpassword", Role: role.User}
	setup.MockRepo.On("GetByUsername", "testuser").Return(mockAccount, nil)
	setup.MockBcrypt.On("CheckPasswordHash", "password", "hashedpassword").Return(true)
	setup.MockJWTGenerator.On("GenerateOriginToken", ctx, mockAccount.ID, mockAccount.Username, role.User, mock.Anything).Return("mockToken", nil)

	redirectURL, err := setup.UseCase.Login(ctx, "testuser", "password")

	assert.NoError(t, err)
	assert.Contains(t, redirectURL, "mockToken")
	setup.MockRepo.AssertExpectations(t)
	setup.MockBcrypt.AssertExpectations(t)
	setup.MockJWTGenerator.AssertExpectations(t)
}

func TestOriginAccountUseCase_Login_InvalidPassword(t *testing.T) {
	setup := setupOriginAccount()
	ctx := context.Background()

	mockAccount := &account.Account{Username: "testuser", Password: "hashedpassword", Role: role.User}
	setup.MockRepo.On("GetByUsername", "testuser").Return(mockAccount, nil)
	setup.MockBcrypt.On("CheckPasswordHash", "wrongpassword", "hashedpassword").Return(false)

	redirectURL, err := setup.UseCase.Login(ctx, "testuser", "wrongpassword")

	assert.Error(t, err)
	assert.Equal(t, innerErrors.ErrPassword.Error(), err.Error())
	assert.Empty(t, redirectURL)
	setup.MockRepo.AssertExpectations(t)
	setup.MockBcrypt.AssertExpectations(t)
}

func TestOriginAccountUseCase_CreateAccount_Success(t *testing.T) {
	setup := setupOriginAccount()
	ctx := context.Background()

	setup.MockBcrypt.On("HashPassword", mock.AnythingOfType("string")).Return("hashedpassword", nil)
	setup.MockRepo.On("Create", mock.AnythingOfType("account.Account")).Return(nil)

	acc, err := setup.UseCase.CreateAccount(ctx, role.Admin, "newuser", role.User)

	assert.NoError(t, err)
	assert.Equal(t, "newuser", acc.Username)
	setup.MockRepo.AssertExpectations(t)
	setup.MockBcrypt.AssertExpectations(t)
}

func TestOriginAccountUseCase_CreateAccount_Unauthorized(t *testing.T) {
	setup := setupOriginAccount()
	ctx := context.Background()

	_, err := setup.UseCase.CreateAccount(ctx, role.User, "newuser", role.User)

	assert.Error(t, err)
	assert.Equal(t, "unauthorized access", err.Error())
}

func TestOriginAccountUseCase_Login_UserNotFound(t *testing.T) {
	setup := setupOriginAccount()
	ctx := context.Background()

	setup.MockRepo.On("GetByUsername", "nonexistentuser").Return((*account.Account)(nil), errors.New("user not found"))

	redirectURL, err := setup.UseCase.Login(ctx, "nonexistentuser", "password")

	assert.Error(t, err)
	assert.Equal(t, innerErrors.ErrPassword.Error(), err.Error())
	assert.Empty(t, redirectURL)
	setup.MockRepo.AssertExpectations(t)
}

func TestOriginAccountUseCase_CreateAccount_UsernameExists(t *testing.T) {
	setup := setupOriginAccount()
	ctx := context.Background()

	// Set up the mock expectations
	setup.MockBcrypt.On("HashPassword", mock.AnythingOfType("string")).Return("hashedpassword", nil)
	setup.MockRepo.On("Create", mock.Anything).Return(innerErrors.ErrExists)

	_, err := setup.UseCase.CreateAccount(ctx, role.Admin, "existinguser", role.User)

	assert.Error(t, err)
	assert.Equal(t, innerErrors.ErrExists.Error(), err.Error())
	setup.MockRepo.AssertExpectations(t)
	setup.MockBcrypt.AssertExpectations(t)
}

func TestOriginAccountUseCase_GetAccountList_Success(t *testing.T) {
	setup := setupOriginAccount()
	ctx := context.Background()

	mockAccounts := []account.Account{
		{Username: "user1", Role: role.User},
		{Username: "user2", Role: role.User},
	}
	setup.MockRepo.On("GetAll").Return(mockAccounts, nil)

	accounts, err := setup.UseCase.GetAccountList(ctx, role.Admin)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(accounts))
	setup.MockRepo.AssertExpectations(t)
}

func TestOriginAccountUseCase_GetAccountList_Unauthorized(t *testing.T) {
	setup := setupOriginAccount()
	ctx := context.Background()

	_, err := setup.UseCase.GetAccountList(ctx, role.User)

	assert.Error(t, err)
	assert.Equal(t, innerErrors.ErrUnauthorized.Error(), err.Error())
}

func TestOriginAccountUseCase_DeleteAccount_Success(t *testing.T) {
	setup := setupOriginAccount()
	ctx := context.Background()

	setup.MockRepo.On("Delete", "user1").Return(nil)

	err := setup.UseCase.DeleteAccount(ctx, role.Admin, "user1")

	assert.NoError(t, err)
	setup.MockRepo.AssertExpectations(t)
}

func TestOriginAccountUseCase_DeleteAccount_Unauthorized(t *testing.T) {
	setup := setupOriginAccount()
	ctx := context.Background()

	err := setup.UseCase.DeleteAccount(ctx, role.User, "user1")

	assert.Error(t, err)
	assert.Equal(t, innerErrors.ErrUnauthorized.Error(), err.Error())
}

func TestOriginAccountUseCase_ChangePassword_Success(t *testing.T) {
	setup := setupOriginAccount()
	ctx := context.Background()

	mockAccount := &account.Account{Username: "user1", Password: "oldhashedpassword"}
	setup.MockRepo.On("GetByUsername", "user1").Return(mockAccount, nil)
	setup.MockBcrypt.On("CheckPasswordHash", "oldpassword", "oldhashedpassword").Return(true)
	setup.MockBcrypt.On("HashPassword", "newpassword").Return("newhashedpassword", nil)
	setup.MockRepo.On("Update", mock.AnythingOfType("account.Account")).Return(nil)

	err := setup.UseCase.ChangePassword(ctx, role.User, "user1", "oldpassword", "newpassword")

	assert.NoError(t, err)
	setup.MockRepo.AssertExpectations(t)
	setup.MockBcrypt.AssertExpectations(t)
}

func TestOriginAccountUseCase_ChangePassword_IncorrectOldPassword(t *testing.T) {
	setup := setupOriginAccount()
	ctx := context.Background()

	mockAccount := &account.Account{Username: "user1", Password: "oldhashedpassword"}
	setup.MockRepo.On("GetByUsername", "user1").Return(mockAccount, nil)
	setup.MockBcrypt.On("CheckPasswordHash", "wrongoldpassword", "oldhashedpassword").Return(false)

	err := setup.UseCase.ChangePassword(ctx, role.User, "user1", "wrongoldpassword", "newpassword")

	assert.Error(t, err)
	assert.Equal(t, innerErrors.ErrPassword.Error(), err.Error())
	setup.MockRepo.AssertExpectations(t)
	setup.MockBcrypt.AssertExpectations(t)
}

func TestOriginAccountUseCase_ChangePassword_Unauthorized(t *testing.T) {
	setup := setupOriginAccount()
	ctx := context.Background()

	err := setup.UseCase.ChangePassword(ctx, role.Guest, "user1", "oldpassword", "newpassword")

	assert.Error(t, err)
	assert.Equal(t, innerErrors.ErrUnauthorized.Error(), err.Error())
}

func TestOriginAccountUseCase_CreateAccount_AdminRoleNotAllowed(t *testing.T) {
	setup := setupOriginAccount()
	ctx := context.Background()

	_, err := setup.UseCase.CreateAccount(ctx, role.Admin, "newadminuser", role.Admin)

	assert.Error(t, err)
	assert.Equal(t, "admin role cannot be assigned to a user", err.Error())
}