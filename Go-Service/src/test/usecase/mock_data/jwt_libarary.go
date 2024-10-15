package mock_data

import (
	"Go-Service/src/main/application/dto"
	"Go-Service/src/main/domain/entity/role"
	"context"

	"github.com/stretchr/testify/mock"
)

type MockJWTGenerator struct {
	mock.Mock
}

// GenerateToken is a mock implementation of the JWTGenerator interface method
func (m *MockJWTGenerator) GenerateToken(ctx context.Context, discordId string, guildMemberData *dto.DiscordGuildMemberDTO, userRole role.Role, secretKey string) (string, error) {
	args := m.Called(ctx, discordId, guildMemberData, userRole, secretKey)
	return args.String(0), args.Error(1)
}