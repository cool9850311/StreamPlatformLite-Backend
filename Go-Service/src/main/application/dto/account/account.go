package dto

import "github.com/cool9850311/StreamPlatformLite-Core/pkg/role"

type AccountListDTO struct {
	Username string    `json:"username"`
	Role     role.Role `json:"role"`
}
