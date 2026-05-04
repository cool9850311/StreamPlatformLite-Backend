package mock_data

import (
	"Go-Service/src/main/application/dto"

	"github.com/stretchr/testify/mock"
)

type MockJWTGenerator struct {
	mock.Mock
}

func (m *MockJWTGenerator) GenerateAnonymousViewerToken(viewerID, secretKey string) (string, error) {
	args := m.Called(viewerID, secretKey)
	return args.String(0), args.Error(1)
}

func (m *MockJWTGenerator) ParseAnonymousViewerToken(tokenString, secretKey string) (*dto.AnonymousViewerClaims, error) {
	args := m.Called(tokenString, secretKey)
	if args.Get(0) != nil {
		return args.Get(0).(*dto.AnonymousViewerClaims), args.Error(1)
	}
	return nil, args.Error(1)
}
