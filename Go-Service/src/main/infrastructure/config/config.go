// Go-Service/src/main/infrastructure/config/config.go
package config

import (
	"log"
	"os"
	"strconv"
	"Go-Service/src/main/infrastructure/util"
	"github.com/joho/godotenv"
)

type Config struct {
	Server struct {
		Port int `mapstructure:"port"`
	} `mapstructure:"server"`
	MongoDB struct {
		URI      string `mapstructure:"uri"`
		Database string `mapstructure:"database"`
	} `mapstructure:"mongodb"`
	JWT struct {
		SecretKey string `mapstructure:"secretKey"`
	} `mapstructure:"JWT"`
}

var AppConfig Config

func LoadConfig() {
	// Load .env file
	projectRootPath, err := util.GetProjectRootPath()
	if err != nil {
		log.Fatalf("Error getting project root path: %s", err)
	}
	err = godotenv.Load(projectRootPath + "/.env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Read environment variables
	port, err := strconv.Atoi(os.Getenv("SERVER_PORT"))
	if err != nil {
		log.Fatalf("Invalid SERVER_PORT: %s", err)
	}
	AppConfig.Server.Port = port
	AppConfig.MongoDB.URI = os.Getenv("MONGODB_URI")
	AppConfig.MongoDB.Database = os.Getenv("MONGODB_DATABASE")
	AppConfig.JWT.SecretKey = os.Getenv("APP_SECRET_KEY")
}
