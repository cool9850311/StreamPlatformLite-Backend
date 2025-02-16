package livestream

// Define the Livestream struct
type Livestream struct {
	UUID        string     `json:"uuid"`
	Name        string     `json:"name"`
	APIKey      string     `json:"api_key"`
	OwnerUserId string     `json:"owner_user_id"`
	Visibility  Visibility `json:"visibility"`
	Title       string     `json:"title"`
	Information string     `json:"information"`
	BanList     []string   `json:"ban_list"`
	MuteList    []string   `json:"mute_list"`
	IsRecord    bool       `json:"is_record"`
}

type Visibility string

const (
	Public     Visibility = "public"
	MemberOnly Visibility = "member_only"
	Private    Visibility = "private"
	Link       Visibility = "link"
)
