package chat

import "github.com/cool9850311/StreamPlatformLite-Core/pkg/role"

type Chat struct {
	ID       string    `json:"id"`
	UserID   string    `json:"user_id"`
	Avatar   string    `json:"avatar"`
	Username string    `json:"username"`
	Message  string    `json:"message"`
	Role     role.Role `json:"role"`
}
