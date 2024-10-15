package controller

import (
	"Go-Service/src/main/domain/interface/logger"
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
	redirectURL, err := c.discordLoginUseCase.Login(ctx, code)
	if err != nil {
		ctx.Redirect(http.StatusFound, redirectURL)
		return
	}
	ctx.Redirect(http.StatusFound, redirectURL)

}
