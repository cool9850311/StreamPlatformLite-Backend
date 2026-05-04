package jwt

import (
	"Go-Service/src/main/application/dto"
)

type JWTGenerator interface {
	GenerateAnonymousViewerToken(viewerID, secretKey string) (string, error)
	ParseAnonymousViewerToken(tokenString, secretKey string) (*dto.AnonymousViewerClaims, error)
}
