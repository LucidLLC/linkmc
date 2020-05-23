package bot

type Context interface {
	User() string
	ChatID() string

	SendMessage(string)

	Bot() Bot
}
