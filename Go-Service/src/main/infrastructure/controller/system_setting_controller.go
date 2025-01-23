package controller

import (
	"Go-Service/src/main/application/dto"
	"Go-Service/src/main/application/usecase"
	"Go-Service/src/main/domain/entity/system"
	"Go-Service/src/main/domain/interface/logger"
	"Go-Service/src/main/infrastructure/message"
	"github.com/gin-gonic/gin"
	"net/http"
)

type SystemSettingController struct {
	Log                  logger.Logger
	systemSettingUseCase *usecase.SystemSettingUseCase
}

func NewSystemSettingController(log logger.Logger, systemSettingUseCase *usecase.SystemSettingUseCase) *SystemSettingController {
	return &SystemSettingController{
		Log:                  log,
		systemSettingUseCase: systemSettingUseCase,
	}
}

func (c *SystemSettingController) GetSetting(ctx *gin.Context) {
	claims := ctx.Request.Context().Value("claims").(*dto.Claims)
	setting, err := c.systemSettingUseCase.GetSetting(ctx, claims.Role)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	ctx.JSON(http.StatusOK, setting)
}

func (c *SystemSettingController) SetSetting(ctx *gin.Context) {
	claims := ctx.Request.Context().Value("claims").(*dto.Claims)
	setting := &system.Setting{}
	if err := ctx.BindJSON(setting); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": message.MsgBadRequest})
		return
	}
	err := c.systemSettingUseCase.SetSetting(ctx, setting, claims.Role)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": message.MsgOK})
}
