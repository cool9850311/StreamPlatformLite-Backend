package controller

import (
	"Go-Service/src/main/domain/interface/logger"
	"Go-Service/src/main/infrastructure/config"
	"fmt"
	"net/http"

	"Go-Service/src/main/application/usecase"

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

func (c *DiscordOauthController) Callback(ctx *gin.Context) {

	code := ctx.Query("code")
	token, redirectURL, err := c.discordLoginUseCase.Login(ctx, code)

	c.Log.Info(ctx, fmt.Sprintf("🔍 DEBUG Callback: token=%s, redirectURL=%s, err=%v", token, redirectURL, err))

	if err != nil {
		c.Log.Error(ctx, fmt.Sprintf("Login error, redirecting to: %s", redirectURL))
		ctx.Redirect(http.StatusFound, redirectURL)
		return
	}

	// Set HttpOnly cookie with token
	// For localhost, use empty domain string (browsers handle this better)
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

	c.Log.Info(ctx, fmt.Sprintf("🔍 DEBUG Redirecting to: %s", redirectURL))

	// Manual redirect to ensure headers are preserved
	ctx.Header("Location", redirectURL)
	ctx.Status(http.StatusFound)

}

func (c *DiscordOauthController) Logout(ctx *gin.Context) {
	// Clear the HttpOnly cookie by setting Max-Age to -1
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

	ctx.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
