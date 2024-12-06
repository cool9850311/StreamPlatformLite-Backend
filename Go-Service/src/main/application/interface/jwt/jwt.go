package jwt

import (
	"Go-Service/src/main/application/dto"
	"Go-Service/src/main/domain/entity/role"
	"context"
)

type JWTGenerator interface {
	GenerateDiscordToken(ctx context.Context, discordId string, guildMemberData *dto.DiscordGuildMemberDTO, userRole role.Role, secretKey string) (string, error)
	GenerateOriginToken(ctx context.Context, username string, userRole role.Role, secretKey string) (string, error)
}
