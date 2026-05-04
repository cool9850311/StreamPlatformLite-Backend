package util

import (
	"Go-Service/src/main/application/dto"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTLibrary struct{}

func NewJWTLibrary() *JWTLibrary {
	return &JWTLibrary{}
}

func (j *JWTLibrary) GenerateAnonymousViewerToken(viewerID, secretKey string) (string, error) {
	cl := dto.AnonymousViewerClaims{
		ViewerID: viewerID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, cl)
	return token.SignedString([]byte(secretKey))
}

func (j *JWTLibrary) ParseAnonymousViewerToken(tokenString, secretKey string) (*dto.AnonymousViewerClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &dto.AnonymousViewerClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(secretKey), nil
		})
	if err != nil {
		return nil, err
	}
	cl, ok := token.Claims.(*dto.AnonymousViewerClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return cl, nil
}
