package usecase

import "Go-Service/src/main/domain/entity"

type SkeletonRepository interface {
	GetByID(id string) (*entity.Skeleton, error)
	Create(skeleton *entity.Skeleton) error
}

type SkeletonUseCase struct {
	SkeletonRepo SkeletonRepository
}

func (u *SkeletonUseCase) GetSkeletonByID(id string) (*entity.Skeleton, error) {
	return u.SkeletonRepo.GetByID(id)
}

func (u *SkeletonUseCase) CreateSkeleton(skeleton *entity.Skeleton) error {
	return u.SkeletonRepo.Create(skeleton)
}
