package controller

import (
	"Go-Service/src/main/application/dto"
	"Go-Service/src/main/application/usecase"
	"Go-Service/src/main/domain/entity/errors"
	"Go-Service/src/main/domain/entity/role"
	"Go-Service/src/main/domain/interface/logger"
	"Go-Service/src/main/infrastructure/config"
	"Go-Service/src/main/infrastructure/message"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type OriginAccountController struct {
	Log                  logger.Logger
	originAccountUseCase *usecase.OriginAccountUseCase
}

func NewOriginAccountController(log logger.Logger, originAccountUseCase *usecase.OriginAccountUseCase) *OriginAccountController {
	return &OriginAccountController{
		Log:                  log,
		originAccountUseCase: originAccountUseCase,
	}
}

// getClaims safely extracts claims from context
func (c *OriginAccountController) getClaims(ctx *gin.Context) (*dto.Claims, error) {
	claimsValue := ctx.Request.Context().Value("claims")
	if claimsValue == nil {
		return nil, errors.ErrUnauthorized
	}

	claims, ok := claimsValue.(*dto.Claims)
	if !ok {
		c.Log.Error(ctx, "Failed to assert claims type")
		return nil, errors.ErrInternal
	}

	return claims, nil
}
func (c *OriginAccountController) Login(ctx *gin.Context) {
	var loginRequest struct {
		Username string `form:"username"`
		Password string `form:"password"`
	}

	if err := ctx.ShouldBind(&loginRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": message.MsgBadRequest})
		return
	}

	if loginRequest.Username == "" || loginRequest.Password == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Username and password are required"})
		return
	}
	token, err := c.originAccountUseCase.Login(ctx, loginRequest.Username, loginRequest.Password)
	if err != nil {
		if err == errors.ErrPassword {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "Incorrect Username or Password"})
			return
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
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

	// Workaround: Gin has a bug where SetCookie doesn't work properly
	// Use manual Set-Cookie header instead
	cookieValue := fmt.Sprintf("token=%s; Path=/; Max-Age=86400; HttpOnly; SameSite=%s%s", token, sameSite, secure)
	if domain != "" {
		cookieValue = fmt.Sprintf("token=%s; Path=/; Domain=%s; Max-Age=86400; HttpOnly; SameSite=%s%s", token, domain, sameSite, secure)
	}

	ctx.Header("Set-Cookie", cookieValue)
	ctx.JSON(http.StatusOK, gin.H{"message": "Login successful"})

}
func (c *OriginAccountController) CreateAccount(ctx *gin.Context) {
	var createAccountRequest struct {
		Username string    `json:"username"`
		Role     role.Role `json:"role"`
	}

	if err := ctx.ShouldBindJSON(&createAccountRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": message.MsgBadRequest})
		return
	}

	if createAccountRequest.Username == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Username is required"})
		return
	}
	claims, err := c.getClaims(ctx)
	if err != nil {
		if err == errors.ErrUnauthorized {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": message.MsgUnauthorized})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	account, err := c.originAccountUseCase.CreateAccount(ctx, claims.Role, createAccountRequest.Username, createAccountRequest.Role)
	if err != nil {
		if err == errors.ErrDuplicate {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "Username already exists"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	ctx.JSON(http.StatusOK, account)
}
func (c *OriginAccountController) GetAccountList(ctx *gin.Context) {
	claims, err := c.getClaims(ctx)
	if err != nil {
		if err == errors.ErrUnauthorized {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": message.MsgUnauthorized})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	accounts, err := c.originAccountUseCase.GetAccountList(ctx, claims.Role)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	ctx.JSON(http.StatusOK, accounts)
}
func (c *OriginAccountController) DeleteAccount(ctx *gin.Context) {
	var deleteAccountRequest struct {
		Username string `json:"username"`
	}

	if err := ctx.ShouldBindJSON(&deleteAccountRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": message.MsgBadRequest})
		return
	}

	if deleteAccountRequest.Username == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Username is required"})
		return
	}

	claims, err := c.getClaims(ctx)
	if err != nil {
		if err == errors.ErrUnauthorized {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": message.MsgUnauthorized})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	if err := c.originAccountUseCase.DeleteAccount(ctx, claims.Role, deleteAccountRequest.Username); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": message.MsgOK})
}

func (c *OriginAccountController) GetMe(ctx *gin.Context) {
	claims, err := c.getClaims(ctx)
	if err != nil {
		if err == errors.ErrUnauthorized {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": message.MsgUnauthorized})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}

	// 返回当前用户信息（包括Anonymous用户）
	// 前端可以通过role字段判断用户类型
	ctx.JSON(http.StatusOK, gin.H{
		"user_id":           claims.UserID,
		"username":          claims.UserName,
		"avatar":            claims.Avatar,
		"role":              claims.Role,
		"identity_provider": claims.IdentityProvider,
	})
}

func (c *OriginAccountController) ChangePassword(ctx *gin.Context) {
	var changePasswordRequest struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := ctx.ShouldBindJSON(&changePasswordRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": message.MsgBadRequest})
		return
	}

	if changePasswordRequest.OldPassword == "" || changePasswordRequest.NewPassword == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Old password and new password are required"})
		return
	}

	claims, err := c.getClaims(ctx)
	if err != nil {
		if err == errors.ErrUnauthorized {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": message.MsgUnauthorized})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	if err := c.originAccountUseCase.ChangePassword(ctx, claims.Role, claims.UserName, changePasswordRequest.OldPassword, changePasswordRequest.NewPassword); err != nil {
		if err == errors.ErrNotFound {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "Account not found"})
			return
		}
		if err == errors.ErrPassword {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "Incorrect old password"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": message.MsgOK})
}
