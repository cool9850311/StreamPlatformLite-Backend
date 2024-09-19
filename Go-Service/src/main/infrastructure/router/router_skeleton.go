// Go-Service/src/main/infrastructure/router/router.go
package router

import (
	"Go-Service/src/main/application/interface/stream"
	"Go-Service/src/main/application/usecase"
	"Go-Service/src/main/domain/interface/logger"
	"Go-Service/src/main/infrastructure/config"
	"Go-Service/src/main/infrastructure/controller"
	"Go-Service/src/main/infrastructure/middleware"
	"Go-Service/src/main/infrastructure/repository"
	"Go-Service/src/main/infrastructure/cache"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"github.com/redis/go-redis/v9"
)

func NewRouter(db *mongo.Database, log logger.Logger, liveStreamService stream.ILivestreamService, redisClient *redis.Client) *gin.Engine {
	r := gin.Default()

	systemSettingRepo := repository.NewMongoSystemSettingRepository(db)
	systemSettingUseCase := usecase.NewSystemSettingUseCase(systemSettingRepo, log)
	systemSettingController := controller.NewSystemSettingController(log, systemSettingUseCase)
	discordLoginUseCase := usecase.NewDiscordLoginUseCase(systemSettingRepo, log, config.AppConfig)
	skeletonRepo := repository.NewMongoSkeletonRepository(db)
	userRepo := repository.NewUserRepository(db)
	skeletonUseCase := &usecase.SkeletonUseCase{SkeletonRepo: skeletonRepo, Log: log}
	skeletonController := &controller.SkeletonController{SkeletonUseCase: skeletonUseCase, Log: log}
	authController := &controller.AuthController{Log: log, UserRepository: userRepo}
	liveStreamHLSController := &controller.LiveStreamHLSController{Log: log}
	discordOauthController := controller.NewDiscordOauthController(log, discordLoginUseCase)
	livestreamRepo := repository.NewMongoLivestreamRepository(db)
	viewerCountCache := cache.NewRedisViewerCount(redisClient)
	chatCache := cache.NewRedisChat(redisClient)
	livestreamUseCase := usecase.NewLivestreamUsecase(livestreamRepo, log, config.AppConfig, liveStreamService, viewerCountCache, chatCache)
	livestreamController := controller.NewLivestreamController(log, livestreamUseCase)

	// Add CORS middleware to allow all origins
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Handle OPTIONS requests globally
	r.Use(func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Status(200)
			return
		}
		c.Next()
	})

	r.Use(middleware.TraceIDMiddleware())
	r.POST("/login", authController.Login)
	r.GET("/oauth/discord", discordOauthController.Callback)
	r.GET("/skeletons/:id", middleware.JWTAuthMiddleware(log), skeletonController.GetSkeleton)
	r.POST("/skeletons", middleware.JWTAuthMiddleware(log), skeletonController.CreateSkeleton)
	r.GET("/livestream/:uuid/:filename", liveStreamHLSController.GetFile)
	r.GET("/system-settings", middleware.JWTAuthMiddleware(log), systemSettingController.GetSetting)
	r.PATCH("/system-settings", middleware.JWTAuthMiddleware(log), systemSettingController.SetSetting)
	r.POST("/livestream", middleware.JWTAuthMiddleware(log), livestreamController.CreateLivestream)
	r.GET("/livestream/owner/:user_id", middleware.JWTAuthMiddleware(log), livestreamController.GetLivestreamByOwnerId)
	r.GET("/livestream/one", middleware.JWTAuthMiddleware(log), livestreamController.GetLivestreamOne)
	r.PATCH("/livestream/:uuid", middleware.JWTAuthMiddleware(log), livestreamController.UpdateLivestream)
	r.DELETE("/livestream/:uuid", middleware.JWTAuthMiddleware(log), livestreamController.DeleteLivestream)
	r.GET("/livestream/ping-viewer-count/:uuid", middleware.JWTAuthMiddleware(log), livestreamController.PingViewerCount)
	r.GET("/livestream/chat/:uuid/:index", middleware.JWTAuthMiddleware(log), livestreamController.GetChat)
	r.POST("/livestream/chat", middleware.JWTAuthMiddleware(log), livestreamController.AddChat)
	r.DELETE("/livestream/chat/:uuid/:chat_id", middleware.JWTAuthMiddleware(log), livestreamController.RemoveViewerCount)
	return r
}
