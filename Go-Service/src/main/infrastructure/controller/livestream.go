package controller

import (
	"Go-Service/src/main/domain/interface/logger"
	"path/filepath"
	"os"
	"fmt"
	"github.com/gin-gonic/gin"
	"Go-Service/src/main/infrastructure/util"
	"net/http"
)

type LiveStreamController struct {
	// LiveStreamUseCase *usecase.LiveStreamUseCase
	Log logger.Logger
}

func (c *LiveStreamController) GetFile(ctx *gin.Context) {
	// uuid := ctx.Param("uuid")
	filename := ctx.Param("filename")
	// claims := ctx.Request.Context().Value("claims").(*middleware.Claims)

	// filePath, err := c.LiveStreamUseCase.GetFilePathByUUIDAndFilename(ctx, uuid, filename, claims.Role)
	// if err != nil {
	// 	if err == errors.ErrUnauthorized {
	// 		ctx.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
	// 		return
	// 	}
	// 	ctx.JSON(http.StatusNotFound, gin.H{"message": "File not found"})
	// 	return
	// }
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Header("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
	ctx.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
	ctx.Header("Access-Control-Allow-Credentials", "true")
	ctx.Header("Cache-Control", "no-cache, no-store, must-revalidate")

	// Handle preflight request
	rootPath, err := util.GetProjectRootPath()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to get file"})
		return
	}
	// Set Content-Length and Content-Type headers
	filePath := filepath.Clean(rootPath + "/hls/" + ctx.Param("uuid") + "/" + filename)
	fileInfo, err := os.Stat(filePath)
	if err == nil {
		ctx.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	}

	if filepath.Ext(filename) == ".m3u8" {
		ctx.Header("Content-Type", "application/vnd.apple.mpegurl") // for .m3u8
	} else if filepath.Ext(filename) == ".ts" {
		ctx.Header("Content-Type", "video/mp2t") // for .ts
	}

	ctx.File(filePath)
}
