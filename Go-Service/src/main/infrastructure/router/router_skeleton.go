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
	userRepo := repository.NewUserRepository(db)
	skeletonUseCase := &usecase.SkeletonUseCase{SkeletonRepo: skeletonRepo, Log: log}
	skeletonController := &controller.SkeletonController{SkeletonUseCase: skeletonUseCase, Log: log}
	authController := &controller.AuthController{Log: log, UserRepository: userRepo}
	r.Use(middleware.TraceIDMiddleware())
	r.POST("/login", authController.Login)
	r.GET("/skeletons/:id", middleware.JWTAuthMiddleware(log), skeletonController.GetSkeleton)
	r.POST("/skeletons", middleware.JWTAuthMiddleware(log), skeletonController.CreateSkeleton)

	return r
}
