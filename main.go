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
	c, err := config.Read(*configFlag)

	if err != nil {
		panic(err)
	}

	a := app.New(c)

	a.Run()
}
