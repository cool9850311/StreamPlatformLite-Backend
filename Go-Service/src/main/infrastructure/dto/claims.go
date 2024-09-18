package dto

import (
	"Go-Service/src/main/domain/entity/role"

	"github.com/dgrijalva/jwt-go"
)

type Claims struct {
	UserID           string
	UserName         string
	Role             role.Role
	IdentityProvider string
	jwt.StandardClaims
}
