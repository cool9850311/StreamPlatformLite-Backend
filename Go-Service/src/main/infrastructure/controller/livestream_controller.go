package controller

import (
	livestreamDTO "Go-Service/src/main/application/dto/livestream"
	"Go-Service/src/main/application/usecase"
	"Go-Service/src/main/domain/entity/chat"
	"Go-Service/src/main/domain/entity/errors"
	"Go-Service/src/main/domain/entity/livestream"
	"Go-Service/src/main/domain/interface/logger"
	"Go-Service/src/main/infrastructure/dto"
	"Go-Service/src/main/infrastructure/message"
	"net/http"
	"os"
	"path/filepath"
	"Go-Service/src/main/infrastructure/util"
	"fmt"
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
	id := ctx.Param("user_id")
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
		if err == errors.ErrUnauthorized {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": message.MsgUnauthorized})
			return
		}
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

func (c *LivestreamController) PingViewerCount(ctx *gin.Context) {
	id := ctx.Param("uuid")
	claims := ctx.Request.Context().Value("claims").(*dto.Claims)
	viewerCount, err := c.livestreamUseCase.PingViewerCount(ctx, claims.Role, id, claims.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"viewer_count": viewerCount})
}
func (c *LivestreamController) GetChat(ctx *gin.Context) {
	id := ctx.Param("uuid")
	indexStr := ctx.Param("index")

	claims := ctx.Request.Context().Value("claims").(*dto.Claims)
	chats, err := c.livestreamUseCase.GetChat(ctx, claims.Role, id, indexStr)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	ctx.JSON(http.StatusOK, chats)
}
func (c *LivestreamController) AddChat(ctx *gin.Context) {
	var chatRequest livestreamDTO.LivestreamAddChatRequestDTO
	if err := ctx.ShouldBindJSON(&chatRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	claims := ctx.Request.Context().Value("claims").(*dto.Claims)
	chat := chat.Chat{
		UserID:   claims.UserID,
		Username: claims.UserName,
		Message:  chatRequest.Message,
	}
	err := c.livestreamUseCase.AddChat(ctx, claims.Role, chatRequest.StreamUUID, chat)
	if err != nil {
		if err == errors.ErrMuteUser {
			ctx.JSON(http.StatusForbidden, gin.H{"message": message.MsgForbidden})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	ctx.JSON(http.StatusOK, "Chat added")
}
func (c *LivestreamController) RemoveViewerCount(ctx *gin.Context) {
	id := ctx.Param("uuid")
	chatID := ctx.Param("chat_id")
	claims := ctx.Request.Context().Value("claims").(*dto.Claims)
	err := c.livestreamUseCase.DeleteChat(ctx, claims.Role, id, chatID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	ctx.JSON(http.StatusOK, "Chat deleted")
}
func (c *LivestreamController) GetDeleteChatIDs(ctx *gin.Context) {
	id := ctx.Param("uuid")
	claims := ctx.Request.Context().Value("claims").(*dto.Claims)
	ids, err := c.livestreamUseCase.GetDeleteChatIDs(ctx, claims.Role, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	ctx.JSON(http.StatusOK, ids)
}
func (c *LivestreamController) MuteUser(ctx *gin.Context) {
	var muteUserRequest livestreamDTO.LivestreamMuteUserRequestDTO
	if err := ctx.ShouldBindJSON(&muteUserRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	claims := ctx.Request.Context().Value("claims").(*dto.Claims)

	err := c.livestreamUseCase.MuteUser(ctx, claims.Role, muteUserRequest.StreamUUID, muteUserRequest.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
}

func (c *LivestreamController) GetFile(ctx *gin.Context) {
	// uuid := ctx.Param("uuid")
	filename := ctx.Param("filename")
	claims := ctx.Request.Context().Value("claims").(*dto.Claims)
	err := c.livestreamUseCase.GetStreamFile(ctx, claims.Role)
	if err != nil {
		if err == errors.ErrUnauthorized {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": message.MsgUnauthorized})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Header("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
	ctx.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
	ctx.Header("Access-Control-Allow-Credentials", "true")
	ctx.Header("Cache-Control", "no-cache, no-store, must-revalidate")

	// Handle preflight request
	rootPath, err := util.GetProjectRootPath()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	// Set Content-Length and Content-Type headers
	filePath := filepath.Clean(rootPath + "/hls/" + ctx.Param("uuid") + "/" + filename)
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "File not found"})
		return
	}

	ctx.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	if filepath.Ext(filename) == ".m3u8" {
		ctx.Header("Content-Type", "application/vnd.apple.mpegurl") // for .m3u8
	} else if filepath.Ext(filename) == ".ts" {
		ctx.Header("Content-Type", "video/mp2t") // for .ts
	}

	ctx.File(filePath)
}