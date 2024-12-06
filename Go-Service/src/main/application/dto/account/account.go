package dto

import "Go-Service/src/main/domain/entity/role"
type AccountListDTO struct {
	Username string    `json:"username"`
	Role     role.Role `json:"role"`
}
