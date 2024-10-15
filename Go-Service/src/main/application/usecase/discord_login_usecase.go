package usecase

import (
	"Go-Service/src/main/application/interface/repository"
	"Go-Service/src/main/domain/entity/role"
	"Go-Service/src/main/domain/interface/logger"
	"Go-Service/src/main/application/dto/config"
	"context"
	"fmt"
	"strconv"
	"Go-Service/src/main/domain/entity/errors"
	"Go-Service/src/main/application/interface/outer_api/discord"
	"Go-Service/src/main/application/interface/jwt"
	"Go-Service/src/main/application/dto"
)

type DiscordLoginUseCase struct {
	systemSettingRepo repository.SystemSettingRepository
	Log               logger.Logger
	config            config.Config
	discordOAuth      discord.DiscordOAuth
	jwtGenerator      jwt.JWTGenerator
}

func NewDiscordLoginUseCase(systemSettingRepo repository.SystemSettingRepository, log logger.Logger, config config.Config, discordOAuth discord.DiscordOAuth, jwtGenerator jwt.JWTGenerator) *DiscordLoginUseCase {
	return &DiscordLoginUseCase{
		systemSettingRepo: systemSettingRepo,
		Log:               log,
		config:            config,
		discordOAuth:      discordOAuth,
		jwtGenerator:      jwtGenerator,
	}
}

func (u *DiscordLoginUseCase) Login(ctx context.Context, code string) (string, error) {
	
	var clientRedirectURL string
	if u.config.Server.HTTPS {
		clientRedirectURL = fmt.Sprintf("https://%s", u.config.Frontend.Domain)
	} else {
		clientRedirectURL = fmt.Sprintf("http://%s:%s", u.config.Frontend.Domain, strconv.Itoa(u.config.Frontend.Port))
	}
	if u.config.Discord.ClientID == "" ||
	u.config.Discord.ClientSecret == "" ||
	u.config.Server.Domain == "" ||
	u.config.Frontend.Domain == "" ||
	u.config.Discord.AdminID == "" ||
	u.config.Discord.GuildID == "" {
		u.Log.Error(ctx, "Incomplete Discord configuration")
		return clientRedirectURL, errors.ErrInternal
	}
	if code == "" {
		u.Log.Error(ctx, "Authorization code not found")
		return clientRedirectURL, errors.ErrInvalidInput
	}
	var scheme string
	var redirectURI string
	if u.config.Server.HTTPS {
		scheme = "https://"
		redirectURI = scheme + u.config.Server.Domain+ "/oauth/discord"
	} else {
		scheme = "http://"
		redirectURI = scheme + u.config.Server.Domain + ":" + strconv.Itoa(u.config.Server.Port) + "/oauth/discord"
	}
	accessToken, err := u.discordOAuth.GetAccessToken(ctx, u.config.Discord.ClientID, u.config.Discord.ClientSecret, code, redirectURI)
	if err != nil {
		u.Log.Error(ctx, "Error getting access token")
		return clientRedirectURL, errors.ErrInternal
	}
	discordGuildMemberData, err := u.discordOAuth.GetGuildMemberData(ctx, accessToken, u.config.Discord.GuildID)
	if err != nil {
		u.Log.Error(ctx, "Error getting user discord id")
		return clientRedirectURL, errors.ErrInternal
	}
	discordId := discordGuildMemberData.User.ID
	userDiscordRoles := discordGuildMemberData.Roles

	// Check for admin role
		if discordId == u.config.Discord.AdminID {
		token, err := u.generateToken(ctx, discordId, discordGuildMemberData, role.Admin)
		if err != nil {
			return clientRedirectURL, err
		}
		return u.generateRedirectURL(token)
	}

	setting, err := u.systemSettingRepo.GetSetting()
	if err != nil {
		u.Log.Error(ctx, "Error getting system setting")
		return clientRedirectURL, errors.ErrInternal
	}

	// Check for editor role
	if contains(userDiscordRoles, setting.EditorRoleId) {
		token, err := u.generateToken(ctx, discordId, discordGuildMemberData, role.Editor)
		if err != nil {
			return clientRedirectURL, err
		}
		return u.generateRedirectURL(token)
	}

	// Check for stream access roles
	if hasIntersection(userDiscordRoles, setting.StreamAccessRoleIds) {
		token, err := u.generateToken(ctx, discordId, discordGuildMemberData, role.User)
		if err != nil {
			return clientRedirectURL, err
		}
		return u.generateRedirectURL(token)
	}
	token, err := u.generateToken(ctx, discordId, discordGuildMemberData, role.Guest)
	if err != nil {
		return clientRedirectURL, err
	}
	return u.generateRedirectURL(token)

}

func (u *DiscordLoginUseCase) generateToken(ctx context.Context, discordId string, discordGuildMemberData *dto.DiscordGuildMemberDTO, userRole role.Role) (string, error) {
	jwt, err := u.jwtGenerator.GenerateToken(ctx, discordId, discordGuildMemberData, userRole, u.config.JWT.SecretKey)
	if err != nil {
		u.Log.Error(ctx, "Error generating JWT")
		return "", errors.ErrInternal
	}
	return jwt, nil
}

func (u *DiscordLoginUseCase) generateRedirectURL(token string) (string, error) {
	var redirectURL string
	if u.config.Server.HTTPS {
		redirectURL = fmt.Sprintf("https://%s?token=%s", u.config.Frontend.Domain, token)
	} else {
		redirectURL = fmt.Sprintf("http://%s:%s?token=%s", u.config.Frontend.Domain, strconv.Itoa(u.config.Frontend.Port), token)
	}
	return redirectURL, nil
}

func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func hasIntersection(slice1, slice2 []string) bool {
	for _, v1 := range slice1 {
		for _, v2 := range slice2 {
			if v1 == v2 {
				return true
			}
		}
	}
	return false
}
