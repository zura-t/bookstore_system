package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	HttpPort             string        `mapstructure:"HTTP_PORT"`
	UsersServiceAddress  string        `mapstructure:"USERS_SERVICE_ADDRESS"`
	DbUrl                string        `mapstructure:"DB_URL"`
	TokenKey             string        `mapstructure:"TOKEN_KEY"`
	AccessTokenDuration  time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	LogLevel             string        `mapstructure:"LOG_LEVEL"`
	Environment          string        `mapstructure:"ENVIRONMENT"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
