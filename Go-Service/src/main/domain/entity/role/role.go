package role

type Role int

const (
	Admin Role = iota
	Streamer
	User
	Guest
)
