package usecase

import (
	"Go-Service/src/main/application/interface/repository"
	"Go-Service/src/main/domain/entity/errors"
	"Go-Service/src/main/domain/entity/role"
	"Go-Service/src/main/domain/entity/system"
	"Go-Service/src/main/domain/interface/logger"
	"context"
)

type SystemSettingUseCase struct {
	systemSettingRepo repository.SystemSettingRepository
	Log               logger.Logger
}

func NewSystemSettingUseCase(systemSettingRepo repository.SystemSettingRepository, log logger.Logger) *SystemSettingUseCase {
	return &SystemSettingUseCase{
		systemSettingRepo: systemSettingRepo,
		Log:               log,
	}
}

func (u *SystemSettingUseCase) CheckRole(userRole role.Role) error {
	if userRole != role.Admin {
		return errors.ErrUnauthorized
	}
	return nil
}
func (u *SystemSettingUseCase) GetSetting(ctx context.Context, userRole role.Role) (*system.Setting, error) {
	if err := u.CheckRole(userRole); err != nil {
		u.Log.Error(ctx, "Unauthorized access to GetSetting")
		return nil, err
	}
	setting, err := u.systemSettingRepo.GetSetting()
	if err != nil {
		u.Log.Error(ctx, "Error getting system setting")
		return nil, err
	}
	return setting, nil
}

// set setting
func (u *SystemSettingUseCase) SetSetting(ctx context.Context, setting *system.Setting, userRole role.Role) error {
	if err := u.CheckRole(userRole); err != nil {
		u.Log.Error(ctx, "Unauthorized access to SetSetting")
		return err
	}
	return u.systemSettingRepo.SetSetting(setting)
}
