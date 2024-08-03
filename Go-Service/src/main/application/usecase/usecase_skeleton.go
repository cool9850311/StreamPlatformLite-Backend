package usecase

import (
	"Go-Service/src/main/domain/entity"
	"Go-Service/src/main/domain/interface/logger"
)

type SkeletonRepository interface {
	GetByID(id string) (*entity.Skeleton, error)
	Create(skeleton *entity.Skeleton) error
}

type SkeletonUseCase struct {
	SkeletonRepo SkeletonRepository
	Log          logger.Logger
}

func (u *SkeletonUseCase) GetSkeletonByID(id string) (*entity.Skeleton, error) {
	return u.SkeletonRepo.GetByID(id)
}

func (u *SkeletonUseCase) CreateSkeleton(skeleton *entity.Skeleton) error {
	return u.SkeletonRepo.Create(skeleton)
}
