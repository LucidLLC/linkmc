package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/rbrick/linkmc/config"
	"log"
)

type TelegramBot struct {
	CommandHandler
	conf config.Bot

	updates tgbotapi.UpdatesChannel

	api *tgbotapi.BotAPI
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
				text := u.Message.Text

				if text[0] == '/' {
					b.CommandHandler.Handle(text[1:])
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
