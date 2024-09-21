package dto

import "time"

// UserDTO represents the user part of the guild member.
type DiscordUserDTO struct {
	AccentColor          *int                         `json:"accent_color"`
	Avatar               string                       `json:"avatar"`
	AvatarDecorationData *DiscordAvatarDecorationData `json:"avatar_decoration_data"`
	Banner               string                       `json:"banner"`
	BannerColor          *string                      `json:"banner_color"`
	Clan                 *string                      `json:"clan"`
	Discriminator        string                       `json:"discriminator"`
	Flags                int                          `json:"flags"`
	GlobalName           string                       `json:"global_name"`
	ID                   string                       `json:"id"`
	PublicFlags          int                          `json:"public_flags"`
	Username             string                       `json:"username"`
	Email                *string                      `json:"email"`
}

// GuildMemberDTO represents a guild member.
type DiscordGuildMemberDTO struct {
	Avatar                     *string        `json:"avatar"`
	Banner                     *string        `json:"banner"`
	Bio                        string         `json:"bio"`
	CommunicationDisabledUntil *time.Time     `json:"communication_disabled_until"`
	Deaf                       bool           `json:"deaf"`
	Flags                      int            `json:"flags"`
	JoinedAt                   time.Time      `json:"joined_at"`
	Mute                       bool           `json:"mute"`
	Nick                       *string        `json:"nick"`
	Pending                    bool           `json:"pending"`
	PremiumSince               *time.Time     `json:"premium_since"`
	Roles                      []string       `json:"roles"`
	UnusualDmActivityUntil     *time.Time     `json:"unusual_dm_activity_until"`
	User                       DiscordUserDTO `json:"user"`
}
type DiscordAvatarDecorationData struct {
	Asset string `json:"asset"`
	SkuID string `json:"sku_id"`
}
