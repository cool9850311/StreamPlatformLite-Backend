package usecase

import (
	"Go-Service/src/main/application/dto"
	"Go-Service/src/main/application/dto/config"
	"Go-Service/src/main/application/interface/jwt"
	"Go-Service/src/main/application/interface/outer_api/discord"
	"Go-Service/src/main/application/interface/repository"
	"Go-Service/src/main/application/interface/state_store"
	"Go-Service/src/main/domain/entity/errors"
	"Go-Service/src/main/domain/entity/role"
	"Go-Service/src/main/domain/interface/logger"
	"context"
	"fmt"
	"net/url"
	"strconv"
)

type DiscordLoginUseCase struct {
	systemSettingRepo repository.SystemSettingRepository
	Log               logger.Logger
	config            config.Config
	discordOAuth      discord.DiscordOAuth
	jwtGenerator      jwt.JWTGenerator
	stateStore        state_store.StateStore
}

func NewDiscordLoginUseCase(
	systemSettingRepo repository.SystemSettingRepository,
	log logger.Logger,
	config config.Config,
	discordOAuth discord.DiscordOAuth,
	jwtGenerator jwt.JWTGenerator,
	stateStore state_store.StateStore,
) *DiscordLoginUseCase {
	return &DiscordLoginUseCase{
		systemSettingRepo: systemSettingRepo,
		Log:               log,
		config:            config,
		discordOAuth:      discordOAuth,
		jwtGenerator:      jwtGenerator,
		stateStore:        stateStore,
	}
}

// InitiateLogin generates OAuth state and builds Discord authorization URL
func (u *DiscordLoginUseCase) InitiateLogin(ctx context.Context) (authURL string, err error) {
	// 1. Generate state
	state, err := u.stateStore.GenerateState(ctx)
	if err != nil {
		u.Log.Error(ctx, "Failed to generate state: "+err.Error())
		return "", errors.ErrInternal
	}

	// 2. Build redirect URI
	var redirectURI string
	if u.config.Server.HTTPS {
		redirectURI = fmt.Sprintf("https://%s/oauth/discord", u.config.Server.Domain)
	} else {
		redirectURI = fmt.Sprintf("http://%s:%s/oauth/discord",
			u.config.Server.Domain,
			strconv.Itoa(u.config.Server.Port))
	}

	// 3. Build Discord authorization URL
	authURL = "https://discord.com/api/oauth2/authorize?" + url.Values{
		"client_id":     {u.config.Discord.ClientID},
		"redirect_uri":  {redirectURI},
		"response_type": {"code"},
		"scope":         {"identify email guilds.members.read"},
		"state":         {state},
	}.Encode()

	u.Log.Info(ctx, fmt.Sprintf("Generated OAuth URL with state: %s", state))

	return authURL, nil
}

// ValidateStateAndLogin validates OAuth state and executes login flow
func (u *DiscordLoginUseCase) ValidateStateAndLogin(ctx context.Context, code string, state string) (token string, successURL string, errorURL string, err error) {
	// Build error redirect URL
	errorURL = u.buildErrorRedirectURL()

	// 1. Validate state parameter (CSRF protection)
	if err := u.stateStore.ValidateState(ctx, state); err != nil {
		u.Log.Error(ctx, "Invalid state: "+err.Error())
		return "", "", errorURL, errors.ErrUnauthorized
	}

	// 2. Execute login flow
	token, successURL, err = u.Login(ctx, code)
	return token, successURL, errorURL, err
}

// buildErrorRedirectURL constructs the error redirect URL
func (u *DiscordLoginUseCase) buildErrorRedirectURL() string {
	if u.config.Server.HTTPS {
		return fmt.Sprintf("https://%s/?error=invalid_state", u.config.Frontend.Domain)
	}
	return fmt.Sprintf("http://%s:%s/?error=invalid_state",
		u.config.Frontend.Domain,
		strconv.Itoa(u.config.Frontend.Port))
}

func (u *DiscordLoginUseCase) Login(ctx context.Context, code string) (token string, redirectURL string, err error) {

	var clientRedirectURL string
	if u.config.Server.HTTPS {
		clientRedirectURL = fmt.Sprintf("https://%s/stream", u.config.Frontend.Domain)
	} else {
		clientRedirectURL = fmt.Sprintf("http://%s:%s/stream", u.config.Frontend.Domain, strconv.Itoa(u.config.Frontend.Port))
	}
	if u.config.Discord.ClientID == "" ||
		u.config.Discord.ClientSecret == "" ||
		u.config.Server.Domain == "" ||
		u.config.Frontend.Domain == "" ||
		u.config.Discord.AdminID == "" ||
		u.config.Discord.GuildID == "" {
		u.Log.Error(ctx, "Incomplete Discord configuration")
		return "", clientRedirectURL, errors.ErrInternal
	}
	if code == "" {
		u.Log.Error(ctx, "Authorization code not found")
		return "", clientRedirectURL, errors.ErrInvalidInput
	}
	var scheme string
	var redirectURI string
	if u.config.Server.HTTPS {
		scheme = "https://"
		redirectURI = scheme + u.config.Server.Domain + "/oauth/discord"
	} else {
		scheme = "http://"
		redirectURI = scheme + u.config.Server.Domain + ":" + strconv.Itoa(u.config.Server.Port) + "/oauth/discord"
	}
	accessToken, err := u.discordOAuth.GetAccessToken(ctx, u.config.Discord.ClientID, u.config.Discord.ClientSecret, code, redirectURI)
	if err != nil {
		u.Log.Error(ctx, "Error getting access token: "+err.Error())
		return "", clientRedirectURL, errors.ErrInternal
	}
	discordGuildMemberData, err := u.discordOAuth.GetGuildMemberData(ctx, accessToken, u.config.Discord.GuildID)
	if err != nil {
		u.Log.Error(ctx, "Error getting user discord id: "+err.Error())
		return "", clientRedirectURL, errors.ErrInternal
	}
	discordId := discordGuildMemberData.User.ID
	userDiscordRoles := discordGuildMemberData.Roles

	// Check for admin role
	if discordId == u.config.Discord.AdminID {
		token, err := u.generateToken(ctx, discordId, discordGuildMemberData, role.Admin)
		if err != nil {
			return "", clientRedirectURL, err
		}
		return token, clientRedirectURL, nil
	}

	setting, err := u.systemSettingRepo.GetSetting()
	if err != nil {
		u.Log.Error(ctx, "Error getting system setting: "+err.Error())
		return "", clientRedirectURL, errors.ErrInternal
	}

	// Check for editor role
	if contains(userDiscordRoles, setting.EditorRoleId) {
		token, err := u.generateToken(ctx, discordId, discordGuildMemberData, role.Editor)
		if err != nil {
			return "", clientRedirectURL, err
		}
		return token, clientRedirectURL, nil
	}

	// Check for stream access roles
	if hasIntersection(userDiscordRoles, setting.StreamAccessRoleIds) {
		token, err := u.generateToken(ctx, discordId, discordGuildMemberData, role.User)
		if err != nil {
			return "", clientRedirectURL, err
		}
		return token, clientRedirectURL, nil
	}
	token, err = u.generateToken(ctx, discordId, discordGuildMemberData, role.Guest)
	if err != nil {
		return "", clientRedirectURL, err
	}
	return token, clientRedirectURL, nil

}

func (u *DiscordLoginUseCase) generateToken(ctx context.Context, discordId string, discordGuildMemberData *dto.DiscordGuildMemberDTO, userRole role.Role) (string, error) {
	jwt, err := u.jwtGenerator.GenerateDiscordToken(ctx, discordId, discordGuildMemberData, userRole, u.config.JWT.SecretKey)
	if err != nil {
		u.Log.Error(ctx, "Error generating JWT: "+err.Error())
		return "", errors.ErrInternal
	}
	return jwt, nil
}

func (u *DiscordLoginUseCase) generateRedirectURL(token string) (string, error) {
	var redirectURL string
	if u.config.Server.HTTPS {
		redirectURL = fmt.Sprintf("https://%s", u.config.Frontend.Domain)
	} else {
		redirectURL = fmt.Sprintf("http://%s:%s", u.config.Frontend.Domain, strconv.Itoa(u.config.Frontend.Port))
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
