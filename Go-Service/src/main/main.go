// Go-Service/src/main.go
package main

import (
	"Go-Service/src/main/infrastructure/config"
	"Go-Service/src/main/infrastructure/initializer"
	"Go-Service/src/main/infrastructure/router"
	"context"
	"fmt"
)

func main() {
	// Load config first so LOG_LEVEL is available
	initializer.InitConfig()
	// Initialize logger with config loaded
	initializer.InitLog()
	logger := initializer.Log
	logger.Info(context.TODO(), "Configuration loaded successfully")
	logger.Info(context.TODO(), "start InitSchema")
	initializer.InitSchema()
	logger.Info(context.TODO(), "start InitPostgresClient")
	initializer.InitPostgresClient()
	logger.Info(context.TODO(), "start InitRedisClient")
	initializer.InitRedisClient()
	logger.Info(context.TODO(), "start InitLiveStreamService")
	initializer.InitLiveStreamService(logger, initializer.GormDB)
	logger.Info(context.TODO(), "start InitCronJob")
	initializer.InitCronJob(logger, initializer.GormDB)
	logger.Info(context.TODO(), "start router")
	r := router.NewRouter(initializer.GormDB, initializer.Log, initializer.LiveStreamService, initializer.RedisClient)
	logger.Info(context.TODO(), "Server starting...")

	serverPort := config.AppConfig.Server.Port
	if err := r.Run(fmt.Sprintf(":%d", serverPort)); err != nil {
		logger.Fatal(context.TODO(), err.Error())
	}
}
