package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Database struct {
		Name string `mapstructure:"name"`
	} `mapstructure:"database"`
}

func NewConfig() *Config {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	configDir := filepath.Join(home, ".config", "task-cli")
	viper.AddConfigPath(configDir)

	viper.SetConfigName("config")
	viper.SetConfigType("toml")

	viper.SetDefault("database.name", "task")

	_ = viper.ReadInConfig()

	var cfg Config
	_ = viper.Unmarshal(&cfg)
	return &cfg
}
