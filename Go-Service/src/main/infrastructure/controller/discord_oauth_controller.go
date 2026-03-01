package controller

import (
	"Go-Service/src/main/application/usecase"
	"Go-Service/src/main/domain/interface/logger"
	"Go-Service/src/main/infrastructure/config"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DiscordOauthController struct {
	Log                 logger.Logger
	discordLoginUseCase *usecase.DiscordLoginUseCase
}

func NewDiscordOauthController(log logger.Logger, discordLoginUseCase *usecase.DiscordLoginUseCase) *DiscordOauthController {
	return &DiscordOauthController{
		Log:                 log,
		discordLoginUseCase: discordLoginUseCase,
	}
}

// InitiateLogin handles OAuth initiation
func (c *DiscordOauthController) InitiateLogin(ctx *gin.Context) {
	// Call UseCase to generate auth URL
	authURL, err := c.discordLoginUseCase.InitiateLogin(ctx)
	if err != nil {
		c.Log.Error(ctx, "Failed to initiate login: "+err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate login"})
		return
	}

	// Redirect to Discord
	ctx.Redirect(http.StatusFound, authURL)
}

// Callback handles OAuth callback
func (c *DiscordOauthController) Callback(ctx *gin.Context) {
	// Extract parameters
	code := ctx.Query("code")
	state := ctx.Query("state")

	// Call UseCase to validate state and login
	token, successURL, errorURL, err := c.discordLoginUseCase.ValidateStateAndLogin(ctx, code, state)

	if err != nil {
		c.Log.Error(ctx, "Login error: "+err.Error())
		ctx.Redirect(http.StatusFound, errorURL)
		return
	}

	c.Log.Info(ctx, fmt.Sprintf("🔍 DEBUG Callback: token=%s, successURL=%s", token, successURL))

	// Set HttpOnly cookie with token
	c.setCookie(ctx, token)

	// Redirect to success URL
	c.Log.Info(ctx, fmt.Sprintf("🔍 DEBUG Redirecting to: %s", successURL))
	ctx.Header("Location", successURL)
	ctx.Status(http.StatusFound)
}

// Logout handles user logout
func (c *DiscordOauthController) Logout(ctx *gin.Context) {
	// Clear the HttpOnly cookie
	c.clearCookie(ctx)

	c.Log.Info(ctx, "User logged out successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// setCookie sets the authentication cookie
func (c *DiscordOauthController) setCookie(ctx *gin.Context, token string) {
	domain := ""
	if config.AppConfig.Server.HTTPS {
		domain = config.AppConfig.Frontend.Domain
	}

	sameSite := "Lax"
	if config.AppConfig.Server.HTTPS {
		sameSite = "Strict"
	}

	secure := ""
	if config.AppConfig.Server.HTTPS {
		secure = "; Secure"
	}

	// Workaround: Gin has a bug where SetCookie doesn't work with Redirect
	// Use manual Set-Cookie header instead
	cookieValue := fmt.Sprintf("token=%s; Path=/; Max-Age=86400; HttpOnly; SameSite=%s%s", token, sameSite, secure)
	if domain != "" {
		cookieValue = fmt.Sprintf("token=%s; Path=/; Domain=%s; Max-Age=86400; HttpOnly; SameSite=%s%s", token, domain, sameSite, secure)
	}

	c.Log.Info(ctx, fmt.Sprintf("🔍 DEBUG Setting cookie header: %s", cookieValue))
	ctx.Header("Set-Cookie", cookieValue)
}

// clearCookie clears the authentication cookie
func (c *DiscordOauthController) clearCookie(ctx *gin.Context) {
	domain := ""
	if config.AppConfig.Server.HTTPS {
		domain = config.AppConfig.Frontend.Domain
	}

	sameSite := "Lax"
	if config.AppConfig.Server.HTTPS {
		sameSite = "Strict"
	}

	secure := ""
	if config.AppConfig.Server.HTTPS {
		secure = "; Secure"
	}

	// Set cookie with Max-Age=-1 to delete it
	cookieValue := fmt.Sprintf("token=; Path=/; Max-Age=-1; HttpOnly; SameSite=%s%s", sameSite, secure)
	if domain != "" {
		cookieValue = fmt.Sprintf("token=; Path=/; Domain=%s; Max-Age=-1; HttpOnly; SameSite=%s%s", domain, sameSite, secure)
	}

	c.Log.Info(ctx, fmt.Sprintf("🔍 DEBUG Logout: Clearing cookie with header: %s", cookieValue))
	ctx.Header("Set-Cookie", cookieValue)
}
