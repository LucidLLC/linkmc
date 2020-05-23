package app

import (
	"github.com/rbrick/linkmc/bot"
	"github.com/rbrick/linkmc/config"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	bolt "go.etcd.io/bbolt"
)

type Application struct {
	conf *config.Config
	wg   sync.WaitGroup

	DB *bolt.DB

	Bots []bot.Bot
}

func (app *Application) Run(init func(app *Application)) {
	app.startBots()

	init(app)

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
			botConfig.Name = k

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

			app.Bots = append(app.Bots, bot)
		}
	}
}

func (app *Application) Shutdown() {
	for _, b := range app.Bots {
		err := b.Close()

		if err != nil {
			log.Println("error closing bot:", err)
		}
	}

	_ = app.DB.Close()
}

func NewApp(conf *config.Config) *Application {
	return &Application{conf: conf, Bots: []bot.Bot{}}
}
