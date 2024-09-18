package controller

import (
	"Go-Service/src/main/domain/interface/logger"
	"Go-Service/src/main/infrastructure/config"
	"Go-Service/src/main/infrastructure/dto"
	"Go-Service/src/main/infrastructure/message"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"fmt"

	"github.com/dgrijalva/jwt-go"

	"Go-Service/src/main/application/usecase"

	"github.com/gin-gonic/gin"
)

type DiscordOauthController struct {
	Log                 logger.Logger
	discordLoginUseCase *usecase.DiscordLoginUseCase
}

func NewDiscordOauthController(log logger.Logger, discordLoginUseCase *usecase.DiscordLoginUseCase) *DiscordOauthController {
	return &DiscordOauthController{
		Log: log,
		discordLoginUseCase: discordLoginUseCase,
	}
}

func (c *DiscordOauthController) Callback(ctx *gin.Context) {
	code := ctx.Query("code")
	if code == "" {
		c.Log.Error(ctx, "Authorization code not found")
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Authorization code not found"})
		return
	}

	// Prepare the request to Discord's token endpoint
	data := url.Values{}
	data.Set("client_id", config.AppConfig.Discord.ClientID)
	data.Set("client_secret", config.AppConfig.Discord.ClientSecret)
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)

	// Check if HTTPS is enabled and construct the redirect URI accordingly
	var scheme string
	if config.AppConfig.Server.HTTPS {
		scheme = "https://"
	} else {
		scheme = "http://"
	}
	redirectURI := scheme + config.AppConfig.Server.Domain + ":" + strconv.Itoa(config.AppConfig.Server.Port) + "/oauth/discord"
	data.Set("redirect_uri", redirectURI)

	// Check if required Discord configuration fields exist
	if config.AppConfig.Discord.ClientID == "" || 
	config.AppConfig.Discord.ClientSecret == "" || 
	config.AppConfig.Server.Domain == "" || 
	config.AppConfig.Frontend.Domain == "" ||
	config.AppConfig.Discord.AdminID == "" || 
	config.AppConfig.Discord.GuildID == "" {
		c.Log.Error(ctx, "Incomplete Discord configuration")
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	req, err := http.NewRequest("POST", "https://discord.com/api/oauth2/token", bytes.NewBufferString(data.Encode()))
	if err != nil {
		c.Log.Error(ctx, "Failed to create request: "+err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.Log.Error(ctx, "Failed to request access token: "+err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Log.Error(ctx, "Failed to read response body: "+err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}

	if resp.StatusCode != http.StatusOK {
		c.Log.Error(ctx, "Failed to get access token: "+string(body))
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}

	// Parse the access token from the response
	var tokenResponse struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		c.Log.Error(ctx, "Failed to parse access token: "+err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}

	// Use the access token to get the current user's guild member data
	guildID := config.AppConfig.Discord.GuildID // Replace with your guild ID
	guildMemberReq, err := http.NewRequest("GET", "https://discord.com/api/users/@me/guilds/"+guildID+"/member", nil)
	if err != nil {
		c.Log.Error(ctx, "Failed to create guild member request: "+err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	guildMemberReq.Header.Set("Authorization", "Bearer "+tokenResponse.AccessToken)

	guildMemberResp, err := client.Do(guildMemberReq)
	if err != nil {
		c.Log.Error(ctx, "Failed to request guild member data: "+err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	defer guildMemberResp.Body.Close()

	guildMemberBody, err := io.ReadAll(guildMemberResp.Body)
	if err != nil {
		c.Log.Error(ctx, "Failed to read guild member response body: "+err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}

	if guildMemberResp.StatusCode != http.StatusOK {
		c.Log.Error(ctx, "Failed to get guild member data: "+string(guildMemberBody))
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}

	// Parse the guild member data using DiscordGuildMemberDTO
	var guildMemberData dto.DiscordGuildMemberDTO
	if err := json.Unmarshal(guildMemberBody, &guildMemberData); err != nil {
		c.Log.Error(ctx, "Failed to parse guild member data: "+err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	discordId := guildMemberData.User.ID
	userDiscordRoles := guildMemberData.Roles
	userRole, err := c.discordLoginUseCase.Login(ctx, discordId, userDiscordRoles)
	if err != nil {
		c.Log.Error(ctx, "Failed to login: "+err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, dto.Claims{
		UserID:           discordId,
		UserName:         guildMemberData.User.GlobalName,
		Role:             userRole,
		IdentityProvider: "Discord",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(), // Token expires in 1 day
		},
	})
	tokenString, err := token.SignedString([]byte(config.AppConfig.JWT.SecretKey))
	if err != nil {
		c.Log.Error(ctx, "Failed to sign token: "+err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	if config.AppConfig.Server.HTTPS {
		redirectURL := fmt.Sprintf("https://%s?token=%s", config.AppConfig.Frontend.Domain, tokenString)
		ctx.Redirect(http.StatusFound, redirectURL)
		return
	} 
	redirectURL := fmt.Sprintf("http://%s:%s?token=%s", config.AppConfig.Frontend.Domain, strconv.Itoa(config.AppConfig.Frontend.Port), tokenString)
	ctx.Redirect(http.StatusFound, redirectURL)

}
