package dto

import "github.com/golang-jwt/jwt/v5"

type AnonymousViewerClaims struct {
	ViewerID string `json:"viewer_id"`
	jwt.RegisteredClaims
}
