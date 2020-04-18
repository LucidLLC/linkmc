package telegram

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/rbrick/linkmc/bot"
	"github.com/rbrick/linkmc/config"
	"log"
)

type Bot struct {
	bot.CommandHandler
	conf config.Bot

	updates tgbotapi.UpdatesChannel

	api *tgbotapi.BotAPI
}

func (b *Bot) Init() error {
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

		fmt.Println("now awaiting telegram updates")

		for u := range updates {
			if u.Message != nil {
				text := u.Message.Text

				if text[0] == '/' {
					b.CommandHandler.Receive(text[1:])
				}
			}
		}
	}(api)

	b.api = api
	return nil
}

func (b *Bot) Close() error {
	b.api.StopReceivingUpdates()
	return nil
}

func (b *Bot) Config() config.Bot {
	return b.conf
}

func NewTelegramBot(config config.Bot) bot.Bot {
	return &Bot{
		CommandHandler: bot.NewCommandHandler(),
		conf:           config,
	}
}
