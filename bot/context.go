package bot

type Context interface {
	User() string
	UserID() int
	ChatID() string

	SendMessage(string)

	Bot() Bot
}
