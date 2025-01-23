package account

import "Go-Service/src/main/domain/entity/role"

type Account struct {
	ID       string    `json:"_id" bson:"_id"`
	Username string    `json:"username"`
	Password string    `json:"password"`
	Role     role.Role `json:"role"`
}
