package usecase

import (
	"Go-Service/src/main/domain/entity"
	"Go-Service/src/main/domain/entity/errors"
	"Go-Service/src/main/domain/entity/role"
	"Go-Service/src/main/domain/interface/logger"
	"context"
	"Go-Service/src/main/application/interface/repository"
)


type SkeletonUseCase struct {
	SkeletonRepo repository.SkeletonRepository
	Log          logger.Logger
}

func (u *SkeletonUseCase) CheckRole(userRole role.Role) error {
	if userRole > role.User {
		return errors.ErrUnauthorized
	}
	return nil
}

func (u *SkeletonUseCase) GetSkeletonByID(ctx context.Context, id string, userRole role.Role) (*entity.Skeleton, error) {
	if err := u.CheckRole(userRole); err != nil {
		u.Log.Error(ctx, "Unauthorized access to GetSkeletonByID")
		return nil, err
	}
	return u.SkeletonRepo.GetByID(id)
}

func (u *SkeletonUseCase) CreateSkeleton(ctx context.Context, skeleton *entity.Skeleton, userRole role.Role) error {
	if err := u.CheckRole(userRole); err != nil {
		u.Log.Error(ctx, "Unauthorized access to CreateSkeleton")
		return err
	}
	return u.SkeletonRepo.Create(skeleton)
}
