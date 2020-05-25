package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/rbrick/linkmc/config"
	"log"
	"strconv"
)

type TelegramBot struct {
	CommandHandler
	conf config.Bot

	updates tgbotapi.UpdatesChannel

	api *tgbotapi.BotAPI
}

type TelegramContext struct {
	owner *TelegramBot

	user   string
	chatID string
}

func (c *TelegramContext) User() string {
	return c.user
}

func (c *TelegramContext) ChatID() string {
	return c.chatID
}

func (c *TelegramContext) SendMessage(s string) {
	id, err := strconv.Atoi(c.chatID)

	if err != nil {
		id = -1
	}

	if id != -1 {
		_, _ = c.owner.api.Send(tgbotapi.NewMessage(int64(id), s))
	}
}

func (c *TelegramContext) Bot() Bot {
	return c.owner
}

func (b *TelegramBot) Init() error {
	log.Println("creating new telegram bot api")
	api, err := tgbotapi.NewBotAPI(b.conf.Token)

	if err != nil {
		return err
	}

	go func(botapi *tgbotapi.BotAPI) {
		updates, err := botapi.GetUpdatesChan(tgbotapi.UpdateConfig{})
		if err != nil {
			panic(err)
		}

		for u := range updates {
			if u.Message != nil {
				if u.Message.From.ID == botapi.Self.ID {
					continue // skip over ourselves
				}

				if u.Message.IsCommand() {
					b.CommandHandler.Handle(u.Message.Command()+" "+u.Message.CommandArguments(), &TelegramContext{
						owner:  b,
						user:   u.Message.Chat.UserName,
						chatID: strconv.FormatInt(u.Message.Chat.ID, 10),
					})
				}
			}
		}
	}(api)

	log.Println("now awaiting telegram updates")

	b.api = api
	return nil
}

func (b *TelegramBot) Close() error {
	b.api.StopReceivingUpdates()
	return nil
}

func (b *TelegramBot) Config() config.Bot {
	return b.conf
}

func NewTelegramBot(config config.Bot, options ...Option) Bot {
	return &TelegramBot{
		CommandHandler: NewCommandHandler(),
		conf:           config,
	}
}
