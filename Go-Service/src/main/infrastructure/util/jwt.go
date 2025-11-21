package util

import (
	"Go-Service/src/main/application/dto"
	"Go-Service/src/main/domain/entity/role"
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTLibrary struct{}

func NewJWTLibrary() *JWTLibrary {
	return &JWTLibrary{}
}

func (j *JWTLibrary) GenerateDiscordToken(ctx context.Context, discordId string, guildMemberData *dto.DiscordGuildMemberDTO, userRole role.Role, secretKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, dto.Claims{
		UserID:           discordId,
		Avatar:           guildMemberData.User.Avatar,
		UserName:         guildMemberData.User.GlobalName,
		Role:             userRole,
		IdentityProvider: "Discord",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)), // Token expires in 1 day
		},
	})
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
func (j *JWTLibrary) GenerateOriginToken(ctx context.Context, userID string, username string, userRole role.Role, secretKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, dto.Claims{
		UserID:           userID,
		UserName:         username,
		IdentityProvider: "Origin",
		Role:             userRole,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)), // Token expires in 1 day
		},
	})
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
