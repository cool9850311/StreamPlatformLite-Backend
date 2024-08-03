// Go-Service/src/main/infrastructure/controller/skeleton_controller.go
package controller

import (
	"Go-Service/src/main/application/usecase"
	"Go-Service/src/main/domain/entity"
	"Go-Service/src/main/domain/interface/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SkeletonController struct {
	SkeletonUseCase *usecase.SkeletonUseCase
	Log             logger.Logger
}

func (c *SkeletonController) GetSkeleton(ctx *gin.Context) {
	id := ctx.Param("id")
	skeleton, err := c.SkeletonUseCase.GetSkeletonByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "Skeleton not found"})
		return
	}
	c.Log.Info(ctx, "Retrieved skeleton with ID "+id)
	ctx.JSON(http.StatusOK, skeleton)
}

func (c *SkeletonController) CreateSkeleton(ctx *gin.Context) {
	var skeleton entity.Skeleton
	if err := ctx.ShouldBindJSON(&skeleton); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	err := c.SkeletonUseCase.CreateSkeleton(&skeleton)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create skeleton"})
		return
	}
	ctx.JSON(http.StatusCreated, skeleton)
}
