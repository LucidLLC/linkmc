package main

import (
	"flag"
	"github.com/rbrick/linkmc/app"
	"github.com/rbrick/linkmc/config"
	bolt "go.etcd.io/bbolt"
	"os"
)

var (
	configFlag = flag.String("config", "config.toml", "Set the config path to use for the application")

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
