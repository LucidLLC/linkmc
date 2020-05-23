package main

import (
	"fmt"
	"github.com/rbrick/linkmc/bot"
)

func verify(bot bot.Context, name string, args []string) {
	bot.SendMessage(fmt.Sprintf("%s has been verified!", bot.User()))
}
