package discord

import (
	"Go-Service/src/main/application/dto"
	"context"
)

type DiscordOAuth interface {
	GetAccessToken(ctx context.Context, clientID string, clientSecret string, code string, redirectURI string) (string, error)
	GetGuildMemberData(ctx context.Context, accessToken string, guildID string) (*dto.DiscordGuildMemberDTO, error)
}
