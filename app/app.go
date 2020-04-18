package app

import (
	"github.com/rbrick/linkmc/bot"
	"github.com/rbrick/linkmc/config"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Application struct {
	conf *config.Config
	wg   sync.WaitGroup

	bots []bot.Bot
}

func (app *Application) Run() {
	app.startBots()

	log.Println("application started. Press Ctrl+C to terminate")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc // wait for a signal

	app.Shutdown()
}

func (app *Application) startBots() {
	log.Println("initializing bots")

	for k, botConfig := range app.conf.Bots {
		if botConfig.Enabled {
			var opts []bot.Option

			if k == "discord" {
				opts = append(opts, bot.WithVerifyChannel(app.conf.Discord.Channel))
			}

			log.Println("creating", k, "bot")
			bot := bot.Create(k, botConfig, opts...)

			log.Println("created bot", k, "; attempting to start...")

			if err := bot.Init(); err != nil {
				log.Println("error initializing bot", k)
				continue
			}

			app.bots = append(app.bots, bot)
		}
	}
}

func (app *Application) Shutdown() {
	for _, b := range app.bots {
		err := b.Close()

		if err != nil {
			log.Println("error closing bot:", err)
		}
	}
}

func NewApp(conf *config.Config) *Application {
	return &Application{conf: conf, bots: []bot.Bot{}}
}
