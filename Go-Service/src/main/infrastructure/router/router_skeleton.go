// Go-Service/src/main/infrastructure/router/router.go
package router

import (
	"Go-Service/src/main/application/usecase"
	"Go-Service/src/main/domain/interface/logger"
	"Go-Service/src/main/infrastructure/controller"
	"Go-Service/src/main/infrastructure/middleware"
	"Go-Service/src/main/infrastructure/repository"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewRouter(db *mongo.Database, log logger.Logger) *gin.Engine {
	r := gin.Default()

	skeletonRepo := repository.NewMongoSkeletonRepository(db)
	skeletonUseCase := &usecase.SkeletonUseCase{SkeletonRepo: skeletonRepo, Log: log}
	skeletonController := &controller.SkeletonController{SkeletonUseCase: skeletonUseCase, Log: log}
	r.Use(middleware.TraceIDMiddleware())
	r.GET("/skeletons/:id", skeletonController.GetSkeleton)
	r.POST("/skeletons", skeletonController.CreateSkeleton)

	return r
}
