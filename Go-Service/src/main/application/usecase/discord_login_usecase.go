package usecase

import (
	"Go-Service/src/main/application/interface/repository"
	"Go-Service/src/main/domain/entity/role"
	"Go-Service/src/main/domain/interface/logger"
	"Go-Service/src/main/application/dto/config"
	"context"
)

type DiscordLoginUseCase struct {
	systemSettingRepo repository.SystemSettingRepository
	Log               logger.Logger
	config            config.Config
}

func NewDiscordLoginUseCase(systemSettingRepo repository.SystemSettingRepository, log logger.Logger, config config.Config) *DiscordLoginUseCase {
	return &DiscordLoginUseCase{
		systemSettingRepo: systemSettingRepo,
		Log:               log,
		config:            config,
	}
}

func (u *DiscordLoginUseCase) Login(ctx context.Context, userDiscordId string, userDiscordRoles []string) (role.Role, error) {
	if userDiscordId == u.config.Discord.AdminID {
		return role.Admin, nil
	}
	setting, err := u.systemSettingRepo.GetSetting()
	if err != nil {
		u.Log.Error(ctx, "Error getting system setting")
		return role.Guest, err
	}
	
	for _, userRole := range userDiscordRoles {
		if userRole == setting.EditorRoleId {
			return role.Editor, nil
		}
	}

	// Check for intersection between StreamAccessRoleIds and userDiscordRoles
	for _, streamRole := range setting.StreamAccessRoleIds {
		for _, userRole := range userDiscordRoles {
			if streamRole == userRole {
				return role.User, nil
			}
		}
	}

	return role.Guest, nil
}
