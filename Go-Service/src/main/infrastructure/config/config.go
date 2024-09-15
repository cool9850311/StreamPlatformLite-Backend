// Go-Service/src/main/infrastructure/config/config.go
package config

import (
	"log"
	"os"
	"strconv"
	"Go-Service/src/main/infrastructure/util"
	"github.com/joho/godotenv"
	"Go-Service/src/main/application/dto/config"
)



var AppConfig config.Config
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
	AppConfig.Discord.ClientID = os.Getenv("DISCORD_CLIENT_ID")
	AppConfig.Discord.ClientSecret = os.Getenv("DISCORD_CLIENT_SECRET")
	AppConfig.Discord.AdminID = os.Getenv("DISCORD_ADMIN_ID")
	AppConfig.Discord.GuildID = os.Getenv("DISCORD_GUILD_ID")
	AppConfig.Domain = os.Getenv("DOMAIN")
	AppConfig.HTTPS, err = strconv.ParseBool(os.Getenv("HTTPS"))
	if err != nil {
		log.Fatalf("Invalid HTTPS: %s", err)
	}
}
