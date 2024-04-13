package util

import (
	"github.com/spf13/viper"
)

type Config struct {
	ZoneID             string `mapstructure:"ZONE_ID"`
	Address            string `mapstructure:"ADDRESS"`
	Type               string `mapstructure:"TYPE"`
	CloudflareApiToken string `mapstructure:"CLOUDFLARE_API_TOKEN"`
}

func LoadConfig(path string, file string) (Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(file)
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return Config{}, err
	}

	var config Config

	if err := viper.Unmarshal(&config); err != nil {
		return Config{}, err
	}

	return config, nil
}
