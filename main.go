package main

import (
	"flag"
	"github.com/rbrick/linkmc/app"
	"github.com/rbrick/linkmc/config"
)

var (
	configFlag = flag.String("config", "config.toml", "Set the config path to use for the application")
)

func main() {
	flag.Parse()
	c, err := config.Read(*configFlag)

	if err != nil {
		panic(err)
	}

	a := app.NewApp(c)

	a.Run()
}
