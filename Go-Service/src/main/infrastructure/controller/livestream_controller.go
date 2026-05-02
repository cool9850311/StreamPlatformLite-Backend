package controller

import (
	"Go-Service/src/main/application/dto"
	livestreamDTO "Go-Service/src/main/application/dto/livestream"
	jwtInterface "Go-Service/src/main/application/interface/jwt"
	"Go-Service/src/main/application/usecase"
	"Go-Service/src/main/domain/entity/chat"
	"Go-Service/src/main/domain/entity/errors"
	"Go-Service/src/main/domain/entity/livestream"
	"Go-Service/src/main/domain/interface/logger"
	"Go-Service/src/main/infrastructure/config"
	"Go-Service/src/main/infrastructure/message"
	"Go-Service/src/main/infrastructure/util"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type LivestreamController struct {
	Log               logger.Logger
	livestreamUseCase *usecase.LivestreamUsecase
	jwtUtil           jwtInterface.JWTGenerator
}

func NewLivestreamController(log logger.Logger, livestreamUseCase *usecase.LivestreamUsecase, jwtUtil jwtInterface.JWTGenerator) *LivestreamController {
	controller := &LivestreamController{
		Log:               log,
		livestreamUseCase: livestreamUseCase,
		jwtUtil:           jwtUtil,
	}

	return controller
}

// getClaims safely extracts claims from context
func (c *LivestreamController) getClaims(ctx *gin.Context) (*dto.Claims, error) {
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
func (c *LivestreamController) GetLivestreamByID(ctx *gin.Context) {
	id := ctx.Param("uuid")
	claims, err := c.getClaims(ctx)
	if err != nil {
		if err == errors.ErrUnauthorized {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": message.MsgUnauthorized})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	livestream, err := c.livestreamUseCase.GetLivestreamByID(ctx, id, claims.Role)
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

func (c *LivestreamController) GetLivestreamByOwnerId(ctx *gin.Context) {
	id := ctx.Param("user_id")
	claims, err := c.getClaims(ctx)
	if err != nil {
		if err == errors.ErrUnauthorized {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": message.MsgUnauthorized})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	livestream, err := c.livestreamUseCase.GetLivestreamByOwnerID(ctx, id, claims.Role)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	ctx.JSON(http.StatusOK, livestream)
}

func (c *LivestreamController) GetLivestreamOne(ctx *gin.Context) {
	claims, err := c.getClaims(ctx)
	if err != nil {
		if err == errors.ErrUnauthorized {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": message.MsgUnauthorized})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	livestream, err := c.livestreamUseCase.GetOne(ctx, claims.Role)
	if err != nil {
		if err == errors.ErrUnauthorized {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": message.MsgUnauthorized})
			return
		}
		if err == errors.ErrNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"message": message.MsgNotFound})
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
	claims, err := c.getClaims(ctx)
	if err != nil {
		if err == errors.ErrUnauthorized {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": message.MsgUnauthorized})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
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
	claims, err := c.getClaims(ctx)
	if err != nil {
		if err == errors.ErrUnauthorized {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": message.MsgUnauthorized})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	err = c.livestreamUseCase.UpdateLivestream(ctx, &livestream, claims.Role)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	ctx.JSON(http.StatusOK, "Livestream updated")
}

func (c *LivestreamController) DeleteLivestream(ctx *gin.Context) {
	id := ctx.Param("uuid")
	claims, err := c.getClaims(ctx)
	if err != nil {
		if err == errors.ErrUnauthorized {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": message.MsgUnauthorized})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	err = c.livestreamUseCase.DeleteLivestream(ctx, id, claims.Role)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Livestream deleted"})
}

func (c *LivestreamController) PingViewerCount(ctx *gin.Context) {
	id := ctx.Param("uuid")
	claims, err := c.getClaims(ctx)
	if err != nil {
		if err == errors.ErrUnauthorized {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": message.MsgUnauthorized})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}

	var anonymousID string

	tokenStr, cookieErr := ctx.Cookie("anonymous_id")
	if cookieErr == nil && tokenStr != "" {
		anonClaims, parseErr := c.jwtUtil.ParseAnonymousViewerToken(tokenStr, config.AppConfig.JWT.SecretKey)
		if parseErr == nil {
			anonymousID = anonClaims.ViewerID
		}
	}

	if anonymousID == "" {
		clientIP := ctx.ClientIP()
		anonymousID = util.GenerateViewerIDFromIP(clientIP, config.AppConfig.JWT.SecretKey)
		newToken, _ := c.jwtUtil.GenerateAnonymousViewerToken(anonymousID, config.AppConfig.JWT.SecretKey)
		sameSite := http.SameSiteStrictMode
		if !config.AppConfig.Server.HTTPS {
			sameSite = http.SameSiteLaxMode
		}
		http.SetCookie(ctx.Writer, &http.Cookie{
			Name:     "anonymous_id",
			Value:    newToken,
			Path:     "/",
			MaxAge:   86400,
			HttpOnly: true,
			Secure:   config.AppConfig.Server.HTTPS,
			SameSite: sameSite,
		})
	}

	viewerCount, err := c.livestreamUseCase.PingViewerCount(ctx, claims.Role, id, claims.UserID, anonymousID)
	if err != nil {
		if err == errors.ErrUnauthorized {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": message.MsgUnauthorized})
			return
		}
		if err == errors.ErrNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"message": message.MsgNotFound})
			return
		}
		if err == errors.ErrInvalidInput {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": message.MsgInvalidInput})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"viewer_count": viewerCount})
}
func (c *LivestreamController) GetChat(ctx *gin.Context) {
	id := ctx.Param("uuid")
	indexStr := ctx.Param("index")

	claims, err := c.getClaims(ctx)
	if err != nil {
		if err == errors.ErrUnauthorized {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": message.MsgUnauthorized})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	chats, err := c.livestreamUseCase.GetChat(ctx, claims.Role, id, indexStr)
	if err != nil {
		if err == errors.ErrUnauthorized {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": message.MsgUnauthorized})
			return
		}
		if err == errors.ErrNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"message": message.MsgNotFound})
			return
		}
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

	claims, err := c.getClaims(ctx)
	if err != nil {
		if err == errors.ErrUnauthorized {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": message.MsgUnauthorized})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}

	chat := chat.Chat{
		UserID:   claims.UserID,
		Avatar:   claims.Avatar,
		Username: claims.UserName,
		Message:  chatRequest.Message,
		Role:     claims.Role,
	}
	err = c.livestreamUseCase.AddChat(ctx, claims.IdentityProvider, claims.Role, chatRequest.StreamUUID, chat)
	if err != nil {
		if err == errors.ErrMuteUser {
			ctx.JSON(http.StatusForbidden, gin.H{"message": message.MsgForbidden})
			return
		}
		if err == errors.ErrUnauthorized {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": message.MsgUnauthorized})
			return
		}
		c.Log.Error(ctx, "Error adding chat: "+err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	ctx.JSON(http.StatusOK, "Chat added")
}
func (c *LivestreamController) RemoveViewerCount(ctx *gin.Context) {
	id := ctx.Param("uuid")
	chatID := ctx.Param("chat_id")
	claims, err := c.getClaims(ctx)
	if err != nil {
		if err == errors.ErrUnauthorized {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": message.MsgUnauthorized})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	err = c.livestreamUseCase.DeleteChat(ctx, claims.Role, claims.UserID, id, chatID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	ctx.JSON(http.StatusOK, "Chat deleted")
}
func (c *LivestreamController) GetDeleteChatIDs(ctx *gin.Context) {
	id := ctx.Param("uuid")
	claims, err := c.getClaims(ctx)
	if err != nil {
		if err == errors.ErrUnauthorized {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": message.MsgUnauthorized})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	ids, err := c.livestreamUseCase.GetDeleteChatIDs(ctx, claims.Role, id)
	if err != nil {
		if err == errors.ErrUnauthorized {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": message.MsgUnauthorized})
			return
		}
		if err == errors.ErrNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"message": message.MsgNotFound})
			return
		}
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
	claims, err := c.getClaims(ctx)
	if err != nil {
		if err == errors.ErrUnauthorized {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": message.MsgUnauthorized})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}

	err = c.livestreamUseCase.MuteUser(ctx, claims.IdentityProvider, claims.Role, claims.UserID, muteUserRequest.StreamUUID, muteUserRequest.ChatID)
	if err != nil {
		if err == errors.ErrUnauthorized {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": message.MsgUnauthorized})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "User muted successfully"})
}

func (c *LivestreamController) GetFile(ctx *gin.Context) {
	uuidStr := ctx.Param("uuid")
	filename := ctx.Param("filename")

	rootPath, err := util.GetProjectRootPath()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
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

	// Pass rootPath (trusted), uuid and filename (external inputs) to usecase
	fileData, err := c.livestreamUseCase.GetFile(ctx, rootPath, uuidStr, filename, claims.Role)
	if err != nil {
		if err == errors.ErrUnauthorized {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": message.MsgUnauthorized})
			return
		}
		if err == errors.ErrInvalidInput {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input"})
			return
		}
		ctx.JSON(http.StatusNotFound, gin.H{"message": "File not found"})
		return
	}
	ctx.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	ctx.Data(http.StatusOK, getContentType(filename), fileData)
}
func (c *LivestreamController) GetRecord(ctx *gin.Context) {
	uuidStr := ctx.Param("uuid")

	rootPath, err := util.GetProjectRootPath()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
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

	// Pass rootPath (trusted) and uuid (external input) to usecase
	fullFilePath, err := c.livestreamUseCase.GetRecord(ctx, rootPath, uuidStr, claims.Role)
	if err != nil {
		if err == errors.ErrInvalidInput {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input"})
			return
		}
		ctx.JSON(http.StatusNotFound, gin.H{"message": "File not found"})
		return
	}
	file, err := os.Open(fullFilePath)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "File not found"})
		return
	}
	defer file.Close()

	// Get file info to determine size
	fileInfo, err := file.Stat()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": message.MsgInternalServerError})
		return
	}

	ctx.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	ctx.Header("Content-Type", "video/mp4")
	ctx.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	filename := filepath.Base(file.Name())
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", util.EncodeRFC5987(filename)))

	_, err = io.Copy(ctx.Writer, file)
	if err != nil {
		// Cannot send JSON response after headers are sent and streaming has started
		// Just log the error and abort the connection
		c.Log.Error(ctx, "Error streaming file: "+err.Error())
		ctx.Abort()
		return
	}
}

func getContentType(filename string) string {
	if filepath.Ext(filename) == ".m3u8" {
		return "application/vnd.apple.mpegurl"
	} else if filepath.Ext(filename) == ".ts" {
		return "video/mp2t"
	}
	return "application/octet-stream"
}
