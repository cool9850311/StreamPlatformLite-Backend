package account

import "Go-Service/src/main/domain/entity/role"
type Account struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     role.Role `json:"role"`
}
