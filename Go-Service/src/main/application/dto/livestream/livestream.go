package dto

import (
	"Go-Service/src/main/domain/entity/livestream"
)

type LivestreamCreateDTO struct {
	Name        string                `json:"name"`
	Visibility  livestream.Visibility `json:"visibility"`
	Title       string                `json:"title"`
	Information string                `json:"information"`
}
type LivestreamCreateResponseDTO struct {
	StreamPushURL string `json:"streamPushURL"`
}
type LivestreamGetOneResponseDTO struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Title       string `json:"title"`
	Information string `json:"information"`
	StreamURL   string `json:"streamURL"`
}
type LivestreamGetByOwnerIDResponseDTO struct {
	UUID          string                `json:"uuid"`
	Name          string                `json:"name"`
	Visibility    livestream.Visibility `json:"visibility"`
	Title         string                `json:"title"`
	Information   string                `json:"information"`
	StreamPushURL string                `json:"streamPushURL"`
	BanList       []string              `json:"ban_list"`
	MuteList      []string              `json:"mute_list"`
}
type LivestreamAddChatRequestDTO struct {
	StreamUUID string `json:"stream_uuid"`
	Message    string `json:"message"`
}
type LivestreamMuteUserRequestDTO struct {
	StreamUUID string `json:"stream_uuid"`
	UserID     string `json:"user_id"`
}
