package dto

import (
	"Go-Service/src/main/domain/entity/role"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID           string    `json:"user_id"`
	Avatar           string    `json:"avatar"`
	UserName         string    `json:"username"`
	Role             role.Role `json:"role"`
	IdentityProvider string    `json:"identity_provider"`
	jwt.RegisteredClaims
}
