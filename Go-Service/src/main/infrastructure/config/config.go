// Go-Service/src/main/infrastructure/config/config.go
package config

import (
	"fmt"
	"log"
	"os"
	"strings"

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
}

var AppConfig Config

func LoadConfig() {
	workingdir, err := os.Getwd()
	if err != nil {
		log.Fatalf("%s", err)
	}
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(trimPathToBase(workingdir, "Go-Service") + "/src/resource")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("Error unmarshaling config, %s", err)
	}
}
func trimPathToBase(path, base string) string {
	index := strings.Index(path, base)
	if index == -1 {
		fmt.Println("Base path not found in the given path")
		return ""
	}

	trimmedPath := path[:index+len(base)]
	return trimmedPath
}
