package main

import (
	"flag"

	"github.com/LucidLLC/linkmc/app"
	"github.com/LucidLLC/linkmc/config"

	"os"

	bolt "go.etcd.io/bbolt"
)

var (
	configFlag = flag.String("config", "conf/config.toml", "Set the config path to use for the application")

	conf *config.Config

	application *app.Application
)

func main() {
	flag.Parse()
	c, err := config.Read(*configFlag)

	if err != nil {
		panic(err)
	}

	conf = c

	db, err := bolt.Open(c.Database.Path, os.ModePerm, &bolt.Options{})

	if err != nil {
		panic(err)
	}

	application = app.NewApp(c)
	application.DB = db

	application.Run(func(app *app.Application) {

		for _, b := range app.Bots {
			if b.Config().Name == "telegram" {
				b.Register("start", verify)
			} else {
				b.Register("verify", verify)
			}
		}
	})
}
