// Go-Service/src/main/infrastructure/router/router.go
package router

import (
	"Go-Service/src/main/application/interface/stream"
	"Go-Service/src/main/application/usecase"
	"Go-Service/src/main/domain/interface/logger"
	"Go-Service/src/main/infrastructure/cache"
	"Go-Service/src/main/infrastructure/config"
	"Go-Service/src/main/infrastructure/controller"
	"Go-Service/src/main/infrastructure/middleware"
	"Go-Service/src/main/infrastructure/outer_api/discord"
	"Go-Service/src/main/infrastructure/repository"
	"Go-Service/src/main/infrastructure/util"
	"fmt"

	// "github.com/gin-contrib/cors"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewRouter(db *mongo.Database, log logger.Logger, liveStreamService stream.ILivestreamService, redisClient *redis.Client) *gin.Engine {
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
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Content-Disposition", "Cache-Control"},
		AllowCredentials: true,
	}))

	r.Use(middleware.TraceIDMiddleware())
}

func setupRoutes(r *gin.Engine, db *mongo.Database, log logger.Logger, liveStreamService stream.ILivestreamService, redisClient *redis.Client) {
	// Initialize repositories, use cases, and controllers
	systemSettingRepo := repository.NewMongoSystemSettingRepository(db)
	systemSettingUseCase := usecase.NewSystemSettingUseCase(systemSettingRepo, log)
	systemSettingController := controller.NewSystemSettingController(log, systemSettingUseCase)
	discordOAuthOuterApi := discord.NewDiscordOAuthImpl(log)
	jwtGenerator := util.NewJWTLibrary()
	bcrypt := util.NewBcryptLibrary()
	stateStore := util.NewRedisStateStore(redisClient)
	discordLoginUseCase := usecase.NewDiscordLoginUseCase(systemSettingRepo, log, config.AppConfig, discordOAuthOuterApi, jwtGenerator, stateStore)
	discordOauthController := controller.NewDiscordOauthController(log, discordLoginUseCase)
	livestreamRepo := repository.NewMongoLivestreamRepository(db)
	viewerCountCache := cache.NewRedisViewerCount(redisClient)
	chatCache := cache.NewRedisChat(redisClient)
	fileCache := cache.NewFileCache()
	ffmpegLibrary := util.NewFfmpegLibrary()
	livestreamUseCase := usecase.NewLivestreamUsecase(livestreamRepo, log, config.AppConfig, liveStreamService, viewerCountCache, chatCache, fileCache, ffmpegLibrary)
	livestreamController := controller.NewLivestreamController(log, livestreamUseCase)
	accountRepo := repository.NewMongoAccountRepository(db)
	originAccountUseCase := usecase.NewOriginAccountUseCase(accountRepo, log, bcrypt, config.AppConfig, jwtGenerator)
	originAccountController := controller.NewOriginAccountController(log, originAccountUseCase)

	login := r.Group("/")
	{
		login.GET("/oauth/discord/init", discordOauthController.InitiateLogin)
		login.GET("/oauth/discord", discordOauthController.Callback)
		login.POST("/logout", discordOauthController.Logout)
	}
	r.GET("/me", middleware.OptionalJWTAuthMiddleware(log), originAccountController.GetMe)

	originAccount := r.Group("/origin-account")
	{
		originAccount.POST("/login", originAccountController.Login)
		originAccount.POST("/create", middleware.JWTAuthMiddleware(log), originAccountController.CreateAccount)
		originAccount.PATCH("/change-password", middleware.JWTAuthMiddleware(log), originAccountController.ChangePassword)
		originAccount.GET("/list", middleware.JWTAuthMiddleware(log), originAccountController.GetAccountList)
		originAccount.DELETE("/delete", middleware.JWTAuthMiddleware(log), originAccountController.DeleteAccount)
	}

	systemSettings := r.Group("/system-settings")
	{
		systemSettings.GET("", middleware.JWTAuthMiddleware(log), systemSettingController.GetSetting)
		systemSettings.PATCH("", middleware.JWTAuthMiddleware(log), systemSettingController.SetSetting)
	}

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
			chat.POST("", middleware.JWTAuthMiddleware(log), livestreamController.AddChat)
			chat.DELETE("/:uuid/:chat_id", middleware.JWTAuthMiddleware(log), livestreamController.RemoveViewerCount)
		}

		// 禁言功能：需要强制JWT
		livestream.POST("/mute-user", middleware.JWTAuthMiddleware(log), livestreamController.MuteUser)
	}
}
