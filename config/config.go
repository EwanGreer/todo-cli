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
	KeyBinds struct {
		Quit   string `mapstructure:"quit"`
		Down   string `mapstructure:"down"`
		Up     string `mapstructure:"up"`
		Enter  string `mapstructure:"enter"`
		Delete string `mapstructure:"delete"`
		Help   string `mapstructure:"help"`
		Add    string `mapstructure:"add"`
	} `mapstructure:"keybinds"`
	Symbols struct {
		Cursor  string `mapstructure:"cursor"`
		Checked string `mapstructure:"checked"`
	} `mapstructure:"symbols"`
	Colors struct {
		Text     string `mapstructure:"text"`
		Selected string `mapstructure:"selected"`
	} `mapstructure:"colors"`
}

func NewConfig() *Config {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	viper.SetConfigName("default")
	viper.SetConfigType("toml")
	viper.AddConfigPath("config")
	_ = viper.ReadInConfig()

	configDir := filepath.Join(home, ".config", "task-cli")
	viper.AddConfigPath(configDir)

	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	_ = viper.MergeInConfig()

	var cfg Config
	_ = viper.Unmarshal(&cfg)
	return &cfg
}
