// Go-Service/src/main.go
package main

import (
	"Go-Service/src/main/infrastructure/config"
	"Go-Service/src/main/infrastructure/initializer"
	"Go-Service/src/main/infrastructure/router"
	"fmt"
	"log"
)

func main() {
	initializer.InitConfig()
	initializer.InitMongoClient()

	r := router.NewRouter(initializer.DB)

	serverPort := config.AppConfig.Server.Port
	if err := r.Run(fmt.Sprintf(":%d", serverPort)); err != nil {
		log.Fatal(err)
	}
}
