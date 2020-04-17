package config

//Redis section of the config
type Redis struct {
	Host string
	KeyPrefix string
	Index int
}

type Bot struct {
	Token string
	Enabled bool
}
