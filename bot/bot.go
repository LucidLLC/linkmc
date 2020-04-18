package bot

import (
	"github.com/rbrick/linkmc/config"
	"strings"
)

type (
	//CommandFunc represents a command func
	CommandFunc func(string, []string)

	//CommandHandler
	CommandHandler interface {
		//Register registers a new command which is called when a command is received
		Register(cmd string, handler CommandFunc)
		//Receive is called when a command is received
		Handle(cmdline string)
	}

	Option func(bot Bot)

	CreateFunc func(bot config.Bot, options ...Option)
)

//registry is a registry for all of the bots we may have
// update this to add a new bot
var registry = map[string]func(config.Bot, ...Option) Bot{
	"telegram": NewTelegramBot,
	"discord":  NewDiscordBot,
}

type BasicCommandHandler struct {
	commands map[string]CommandFunc
}

func (h *BasicCommandHandler) Register(cmd string, handler CommandFunc) {
	h.commands[cmd] = handler
}

func (h *BasicCommandHandler) Handle(cmdline string) {
	split := strings.Split(cmdline, " ")

	// if the command is present
	if val, ok := h.commands[split[0]]; ok {
		// run the command!
		val(split[0], split[1:])
	}
}

func NewCommandHandler() CommandHandler {
	return &BasicCommandHandler{commands: map[string]CommandFunc{}}
}

func Create(botName string, botConf config.Bot, options ...Option) Bot {
	return registry[botName](botConf, options...)
}

//Bot represents a bot
type Bot interface {
	CommandHandler

	//Init is called to initialize the bot
	Init() error
	//Close is called to close the bot
	Close() error

	// The config for the bot
	Config() config.Bot
}
