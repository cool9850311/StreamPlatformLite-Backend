package role

type Role int

const (
	Admin     Role = iota // 0
	Streamer              // 1
	Editor                // 2
	User                  // 3
	Guest                 // 4
	Anonymous             // 5
)

func (r Role) String() string {
	switch r {
	case Admin:
		return "Admin"
	case Streamer:
		return "Streamer"
	case Editor:
		return "Editor"
	case User:
		return "User"
	case Guest:
		return "Guest"
	case Anonymous:
		return "Anonymous"
	default:
		return "Unknown"
	}
}
