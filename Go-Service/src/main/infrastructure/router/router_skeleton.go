// Go-Service/src/main/infrastructure/router/router.go
package router

import (
	"Go-Service/src/main/application/usecase"
	"Go-Service/src/main/domain/interface/logger"
	"Go-Service/src/main/infrastructure/controller"
	"Go-Service/src/main/infrastructure/middleware"
	"Go-Service/src/main/infrastructure/repository"
	"Go-Service/src/main/application/interface/stream"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewRouter(db *mongo.Database, log logger.Logger, liveStreamService stream.ILivestreamService) *gin.Engine {
	r := gin.Default()

	skeletonRepo := repository.NewMongoSkeletonRepository(db)
	userRepo := repository.NewUserRepository(db)
	skeletonUseCase := &usecase.SkeletonUseCase{SkeletonRepo: skeletonRepo, Log: log}
	skeletonController := &controller.SkeletonController{SkeletonUseCase: skeletonUseCase, Log: log}
	authController := &controller.AuthController{Log: log, UserRepository: userRepo}
	liveStreamController := &controller.LiveStreamController{Log: log} 

	// Add CORS middleware to allow all origins
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Handle OPTIONS requests globally
	r.Use(func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Status(200)
			return
		}
		c.Next()
	})

	r.Use(middleware.TraceIDMiddleware())
	r.POST("/login", authController.Login)
	r.GET("/skeletons/:id", middleware.JWTAuthMiddleware(log), skeletonController.GetSkeleton)
	r.POST("/skeletons", middleware.JWTAuthMiddleware(log), skeletonController.CreateSkeleton)
	r.GET("/livestream/:uuid/:filename", liveStreamController.GetFile) // Added route for LiveStreamController

	return r
}
