package config

//Redis section of the config
type Redis struct {
	Host      string
	KeyPrefix string
	Index     int
}

//Discord is discord specific information
type Discord struct {
	Channel string
}

type Bot struct {
	Token   string
	Enabled bool
}

type Bots struct {
	Bots []Bot
}

type Web struct {
	Host         string
	CallbackPath string
}

type Config struct {
	Bots    Bots
	Discord Discord
	Redis   Redis
	Web     Web
}

func Read(file string) (*Config, error) {
	return nil, nil
}
