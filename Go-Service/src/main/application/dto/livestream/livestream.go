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
