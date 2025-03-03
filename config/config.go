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

	configDir := filepath.Join(home, ".config", "task-cli")
	viper.AddConfigPath(configDir)

	viper.SetConfigName("config")
	viper.SetConfigType("toml")

	// TODO: default config should come from a file
	viper.SetDefault("database.name", "task")

	viper.SetDefault("keybinds.quit", "q")
	viper.SetDefault("keybinds.down", "j")
	viper.SetDefault("keybinds.up", "k")
	viper.SetDefault("keybinds.confirm", "enter")
	viper.SetDefault("keybinds.delete", "d")
	viper.SetDefault("keybinds.help", "?")
	viper.SetDefault("keybinds.add", "a")

	viper.SetDefault("symbols.cursor", ">")
	viper.SetDefault("symbols.checked", "x")

	_ = viper.ReadInConfig()

	var cfg Config
	_ = viper.Unmarshal(&cfg)
	return &cfg
}
