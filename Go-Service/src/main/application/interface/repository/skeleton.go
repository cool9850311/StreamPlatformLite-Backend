package repository

import "Go-Service/src/main/domain/entity"

type SkeletonRepository interface {
	GetByID(id string) (*entity.Skeleton, error)
	Create(skeleton *entity.Skeleton) error
}