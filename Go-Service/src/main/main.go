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
	initializer.InitLog()
	logger := initializer.Log
	logger.Info(context.TODO(), "start InitConfig")
	initializer.InitConfig()
	logger.Info(context.TODO(), "start InitMongoClient")
	initializer.InitMongoClient()
	logger.Info(context.TODO(), "start InitRedisClient")
	initializer.InitRedisClient()
	logger.Info(context.TODO(), "start InitLiveStreamService")
	initializer.InitLiveStreamService(logger, initializer.DB) 
	logger.Info(context.TODO(), "start InitCronJob")
	initializer.InitCronJob(logger, initializer.DB)
	logger.Info(context.TODO(), "start router")
	r := router.NewRouter(initializer.DB, initializer.Log, initializer.LiveStreamService, initializer.RedisClient)
	logger.Info(context.TODO(), "Server starting...")

	serverPort := config.AppConfig.Server.Port
	if err := r.Run(fmt.Sprintf(":%d", serverPort)); err != nil {
		logger.Fatal(context.TODO(), err.Error())
	}
}
