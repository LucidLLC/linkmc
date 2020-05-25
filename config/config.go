package config

import (
	"github.com/BurntSushi/toml"
)

//Redis section of the config
type Redis struct {
	Host      string
	KeyPrefix string `toml:"key_prefix"`
	Index     int
}

//Discord is discord specific information
type Discord struct {
	Channel string
}

type Bot struct {
	Name    string
	Token   string
	Enabled bool
}

type Bots struct {
	Bots []Bot
}

type Twitter struct {
	APIKey    string `toml:"api_key"`
	APISecret string `toml:"api_secret"`
}

type Web struct {
	Host         string
	AuthToken    string `toml:"auth_token"`
	CallbackPath string `toml:"callback_path"`
}

type Database struct {
	Path string
}

type Config struct {
	Bots     map[string]Bot
	Discord  Discord
	Redis    Redis
	Twitter  Twitter
	Web      Web
	Database Database
}

func Read(file string) (*Config, error) {
	var config Config
	if _, err := toml.DecodeFile(file, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
