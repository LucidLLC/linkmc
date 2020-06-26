package app

import (
	"github.com/LucidLLC/linkmc/bot"
	"github.com/LucidLLC/linkmc/config"
	"github.com/LucidLLC/linkmc/web"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	bolt "go.etcd.io/bbolt"
)

type Application struct {
	conf       *config.Config
	wg         sync.WaitGroup
	webHandler *web.Handler
	DB         *bolt.DB

	Bots []bot.Bot
}

func (app *Application) Run(init func(app *Application)) {
	app.startBots()

	init(app)

	app.webHandler = web.New(app.DB, app.conf.Web)

	app.webHandler.Start(app.conf.Web.Host)

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
			} else if k == "telegram" {
				opts = append(opts, bot.WithBroadcastChannel(app.conf.Telegram.Channel))
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
	log.Println("shutting down bots...")
	for _, b := range app.Bots {
		log.Printf("shutting down bot %s", b.Config().Name)
		err := b.Close()

		if err != nil {
			log.Println("error closing bot:", err)
		}
	}

	log.Println("saving database...")
	_ = app.DB.Close()

	if err := app.webHandler.Shutdown(); err != nil {
		log.Panicln(err)
	}
}

func NewApp(conf *config.Config) *Application {
	return &Application{conf: conf, Bots: []bot.Bot{}}
}
