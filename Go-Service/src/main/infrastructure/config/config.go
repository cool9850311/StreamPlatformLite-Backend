// Go-Service/src/main/infrastructure/config/config.go
package config

import (
	"Go-Service/src/main/application/dto/config"
	"Go-Service/src/main/infrastructure/util"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

var AppConfig config.Config

// getEnvAsBool reads an environment variable as a boolean with a default value
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		log.Printf("Invalid value for %s: %s, using default: %t", key, err, defaultValue)
		return defaultValue
	}
	return value
}

// getEnvAsInt64 reads an environment variable as int64 with a default value
func getEnvAsInt64(key string, defaultValue int64) int64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		log.Printf("Invalid value for %s: %s, using default: %d", key, err, defaultValue)
		return defaultValue
	}
	return value
}

func LoadConfig() {
	// Load .env file
	projectRootPath, err := util.GetProjectRootPath()
	if err != nil {
		log.Fatalf("Error getting project root path: %s", err)
	}
	err = godotenv.Load(projectRootPath + "/.env")
	if err != nil {
		log.Println("Error loading .env file")
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
	AppConfig.Server.Domain = os.Getenv("DOMAIN")
	AppConfig.Frontend.Domain = os.Getenv("FRONTEND_DOMAIN")
	AppConfig.Frontend.Port = int(getEnvAsInt64("FRONTEND_PORT", 3000))
	AppConfig.Redis.URI = os.Getenv("REDIS_URI")
	AppConfig.Server.EnableGinLog, err = strconv.ParseBool(os.Getenv("ENABLE_GIN_LOG"))
	if err != nil {
		log.Printf("Invalid ENABLE_GIN_LOG: %s", err)
	}
	AppConfig.Server.HTTPS, err = strconv.ParseBool(os.Getenv("HTTPS"))
	if err != nil {
		log.Printf("Invalid HTTPS: %s", err)
	}

	// Load log level, default to INFO if not set
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "INFO"
	}
	AppConfig.Server.LogLevel = logLevel

	// Load Rate Limiting configuration
	AppConfig.RateLimit.Enabled = getEnvAsBool("RATE_LIMIT_ENABLED", true)
	AppConfig.RateLimit.LoginPerMinute = getEnvAsInt64("RATE_LIMIT_LOGIN_PER_MINUTE", 5)
	AppConfig.RateLimit.OAuthInitPerMinute = getEnvAsInt64("RATE_LIMIT_OAUTH_INIT_PER_MINUTE", 5)
	AppConfig.RateLimit.LogoutPerMinute = getEnvAsInt64("RATE_LIMIT_LOGOUT_PER_MINUTE", 10)
	AppConfig.RateLimit.ChatPostPerMinute = getEnvAsInt64("RATE_LIMIT_CHAT_POST_PER_MINUTE", 10)
	AppConfig.RateLimit.ChatDeletePerMinute = getEnvAsInt64("RATE_LIMIT_CHAT_DELETE_PER_MINUTE", 10)
	AppConfig.RateLimit.ChangePasswordPerHour = getEnvAsInt64("RATE_LIMIT_CHANGE_PASSWORD_PER_HOUR", 10)
}
