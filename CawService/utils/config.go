package utils

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type AppConfig struct {
	Address               string
	Mongo                 string
	TokenExpiresInMinutes int
}

func New() *AppConfig {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	return readConfig()
}

func (appConfig *AppConfig) WithWatchConfig(onChange func(appConfig *AppConfig)) *AppConfig {
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		logger := logrus.New()
		*appConfig = *readConfig()
		onChange(appConfig)
		logger.Println("Config file changed:", appConfig)
	})

	return appConfig
}

func readConfig() *AppConfig {
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	return &AppConfig{
		Address: viper.GetString("address"),
		Mongo:   viper.GetString("mongo"),
		TokenExpiresInMinutes: viper.GetInt("token_expires_in_minutes"),
	}
}
