package controller

import (
	livestreamDTO "Go-Service/src/main/application/dto/livestream"
	"Go-Service/src/main/application/usecase"
	"Go-Service/src/main/domain/entity/errors"
	"Go-Service/src/main/domain/entity/livestream"
	"Go-Service/src/main/domain/interface/logger"
	"Go-Service/src/main/infrastructure/dto"
	"Go-Service/src/main/infrastructure/message"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LivestreamController struct {
	Log               logger.Logger
	livestreamUseCase *usecase.LivestreamUsecase
}

func NewLivestreamController(log logger.Logger, livestreamUseCase *usecase.LivestreamUsecase) *LivestreamController {
	return &LivestreamController{
		Log:               log,
		livestreamUseCase: livestreamUseCase,
	}
}

func (c *LivestreamController) GetLivestreamByOwnerId(ctx *gin.Context) {
	id := ctx.Param("uuid")
	claims := ctx.Request.Context().Value("claims").(*dto.Claims)
	livestream, err := c.livestreamUseCase.GetLivestreamByOwnerID(ctx, id, claims.Role)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	ctx.JSON(http.StatusOK, livestream)
}

func (c *LivestreamController) GetLivestreamOne(ctx *gin.Context) {
	claims := ctx.Request.Context().Value("claims").(*dto.Claims)
	livestream, err := c.livestreamUseCase.GetOne(ctx, claims.Role)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	ctx.JSON(http.StatusOK, livestream)
}

func (c *LivestreamController) CreateLivestream(ctx *gin.Context) {
	var livestreamCreateDTO livestreamDTO.LivestreamCreateDTO
	if err := ctx.ShouldBindJSON(&livestreamCreateDTO); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	claims := ctx.Request.Context().Value("claims").(*dto.Claims)
	livestreamResponse, err := c.livestreamUseCase.CreateLivestream(ctx, &livestreamCreateDTO, claims.UserID, claims.Role)
	if err != nil {
		if err == errors.ErrExists {
			ctx.JSON(http.StatusConflict, gin.H{"message": message.MsgAlreadyExists})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	ctx.JSON(http.StatusOK, livestreamResponse)
}

func (c *LivestreamController) UpdateLivestream(ctx *gin.Context) {
	var livestream livestream.Livestream
	if err := ctx.ShouldBindJSON(&livestream); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	claims := ctx.Request.Context().Value("claims").(*dto.Claims)
	err := c.livestreamUseCase.UpdateLivestream(ctx, &livestream, claims.Role)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	ctx.JSON(http.StatusOK, "Livestream updated")
}

func (c *LivestreamController) DeleteLivestream(ctx *gin.Context) {
	id := ctx.Param("uuid")
	claims := ctx.Request.Context().Value("claims").(*dto.Claims)
	err := c.livestreamUseCase.DeleteLivestream(ctx, id, claims.Role)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Livestream deleted"})
}
