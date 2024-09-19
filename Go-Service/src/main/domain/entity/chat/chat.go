package chat

type Chat struct {
	ID       string `json:"id"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Message  string `json:"message"`
}
