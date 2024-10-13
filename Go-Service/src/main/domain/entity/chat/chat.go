package chat

type Chat struct {
	ID       string `json:"id"`
	UserID   string `json:"user_id"`
	Avatar   string `json:"avatar"`
	Username string `json:"username"`
	Message  string `json:"message"`
}
