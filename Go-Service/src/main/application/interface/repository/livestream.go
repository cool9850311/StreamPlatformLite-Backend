package repository

import "Go-Service/src/main/domain/entity/livestream"

type LivestreamRepository interface {
	GetByID(id string) (*livestream.Livestream, error)
	GetByOwnerID(ownerID string) (*livestream.Livestream, error)
	GetOne() (*livestream.Livestream, error)
	Create(livestream *livestream.Livestream) error
	Update(livestream *livestream.Livestream) error
	Delete(id string) error
}
