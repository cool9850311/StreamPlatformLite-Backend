// Go-Service/src/main/infrastructure/config/config.go
package config

import (
	"Go-Service/src/main/infrastructure/util"
	"log"
	"os"

	"github.com/spf13/viper"
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
	workingdir, err := os.Getwd()
	if err != nil {
		log.Fatalf("%s", err)
	}
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(util.TrimPathToBase(workingdir, "Go-Service") + "/src/resource")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("Error unmarshaling config, %s", err)
	}
}
