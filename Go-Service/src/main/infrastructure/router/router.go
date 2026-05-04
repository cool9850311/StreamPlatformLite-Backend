// Go-Service/src/main/infrastructure/router/router.go
package router

import (
	"Go-Service/src/main/application/interface/stream"
	"Go-Service/src/main/application/usecase"
	"Go-Service/src/main/domain/interface/logger"
	"Go-Service/src/main/infrastructure/cache"
	"Go-Service/src/main/infrastructure/config"
	"Go-Service/src/main/infrastructure/controller"
	"Go-Service/src/main/infrastructure/initializer"
	"Go-Service/src/main/infrastructure/middleware"
	"Go-Service/src/main/infrastructure/repository"
	"Go-Service/src/main/infrastructure/util"
	"fmt"
	"net/http"

	// "github.com/gin-contrib/cors"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB, log logger.Logger, liveStreamService stream.ILivestreamService, redisClient *redis.Client) *gin.Engine {
	r := setupRouter()
	setupMiddlewares(r)
	setupRoutes(r, db, log, liveStreamService, redisClient)
	return r
}

func setupRouter() *gin.Engine {
	if !config.AppConfig.Server.EnableGinLog {
		r := gin.New()
		r.Use(gin.Recovery())
		return r
	}
	return gin.Default()
}

func setupMiddlewares(r *gin.Engine) {
	// Dynamic CORS configuration based on environment
	var allowedOrigins []string
	if config.AppConfig.Server.HTTPS {
		allowedOrigins = append(allowedOrigins, fmt.Sprintf("https://%s", config.AppConfig.Frontend.Domain))
	} else {
		allowedOrigins = append(allowedOrigins, fmt.Sprintf("http://%s:%d", config.AppConfig.Frontend.Domain, config.AppConfig.Frontend.Port))
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-XSRF-TOKEN"},
		ExposeHeaders:    []string{"Content-Length", "Content-Disposition", "Cache-Control", "Retry-After", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"},
		AllowCredentials: true,
	}))

	// 添加安全响应头中间件
	r.Use(middleware.SecurityHeaders())

	r.Use(middleware.TraceIDMiddleware())
}

func setupRoutes(r *gin.Engine, db *gorm.DB, log logger.Logger, liveStreamService stream.ILivestreamService, redisClient *redis.Client) {
	// Initialize rate limiters
	initializer.InitRateLimiters()

	// Initialize repositories, use cases, and controllers
	jwtGenerator := util.NewJWTLibrary()
	livestreamRepo := repository.NewPostgresLivestreamRepository(db)
	viewerCountCache := cache.NewRedisViewerCount(redisClient)
	chatCache := cache.NewRedisChat(redisClient)
	fileCache := cache.NewFileCache()
	ffmpegLibrary := util.NewFfmpegLibrary()
	livestreamUseCase := usecase.NewLivestreamUsecase(livestreamRepo, log, config.AppConfig, liveStreamService, viewerCountCache, chatCache, fileCache, ffmpegLibrary)
	livestreamController := controller.NewLivestreamController(log, livestreamUseCase, jwtGenerator)

	// Health check — public, no auth, used by Docker HEALTHCHECK
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	livestream := r.Group("/livestream")
	{
		// 观看相关端点：使用OptionalJWT中间件（允许匿名访问public直播）
		livestream.GET("/:uuid/:filename", middleware.OptionalJWTAuthMiddleware(log), livestreamController.GetFile)
		livestream.GET("/one", middleware.OptionalJWTAuthMiddleware(log), livestreamController.GetLivestreamOne)
		livestream.GET("/:uuid", middleware.OptionalJWTAuthMiddleware(log), livestreamController.GetLivestreamByID)
		livestream.GET("/ping-viewer-count/:uuid", middleware.OptionalJWTAuthMiddleware(log), livestreamController.PingViewerCount)

		// 管理端点：保持强制JWT（需要Admin权限）
		livestream.POST("", middleware.JWTAuthMiddleware(log), livestreamController.CreateLivestream)
		livestream.PATCH("/:uuid", middleware.JWTAuthMiddleware(log), livestreamController.UpdateLivestream)
		livestream.DELETE("/:uuid", middleware.JWTAuthMiddleware(log), livestreamController.DeleteLivestream)
		livestream.GET("/owner/:user_id", middleware.JWTAuthMiddleware(log), livestreamController.GetLivestreamByOwnerId)
		livestream.GET("/record/:uuid", middleware.JWTAuthMiddleware(log), livestreamController.GetRecord)

		chat := livestream.Group("/chat")
		{
			// 读取聊天：使用OptionalJWT（匿名可读取public直播的聊天）
			chat.GET("/:uuid/:index", middleware.OptionalJWTAuthMiddleware(log), livestreamController.GetChat)
			chat.GET("/delete/:uuid", middleware.OptionalJWTAuthMiddleware(log), livestreamController.GetDeleteChatIDs)

			// 发送/删除聊天：需要强制JWT（需要登录）
			chat.POST("", middleware.JWTAuthMiddleware(log), middleware.RateLimitByUserID(initializer.ChatPostLimiter), livestreamController.AddChat)
			chat.DELETE("/:uuid/:chat_id", middleware.JWTAuthMiddleware(log), middleware.RateLimitByUserID(initializer.ChatDeleteLimiter), livestreamController.RemoveViewerCount)
		}

		// 禁言功能：需要强制JWT
		livestream.POST("/mute-user", middleware.JWTAuthMiddleware(log), livestreamController.MuteUser)
	}
}
