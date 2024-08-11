package entity

import (
	"Go-Service/src/main/domain/entity/role"
	"time"
)

type User struct {
	ID        string
	Username  string
	Password  string
	Role      role.Role
	CreatedAt time.Time
	UpdatedAt time.Time
}