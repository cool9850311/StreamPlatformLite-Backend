package util

import (
	"Go-Service/src/main/application/dto"
	"Go-Service/src/main/domain/entity/role"
	"context"
	"time"
	"github.com/dgrijalva/jwt-go"
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
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(), // Token expires in 1 day
		},
	})
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
func (j *JWTLibrary) GenerateOriginToken(ctx context.Context, username string, userRole role.Role, secretKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, dto.Claims{
		UserName:         username,
		IdentityProvider: "Origin",
		Role:             userRole,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(), // Token expires in 1 day
		},
	})
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
