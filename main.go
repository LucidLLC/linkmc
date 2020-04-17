package main

import (
	"fmt"
	"github.com/rbrick/linkmc/bot/telegram"
	"github.com/rbrick/linkmc/config"
	"sync"
)

func main() {

	bot := telegram.NewTelegramBot(config.Bot{Token: "984860470:AAHlUnN-IuQJX0JM6xyac0s3152WBQBwoSY"})

	bot.Register("start", func(name string, args []string) {

		fmt.Println("received command start")
	})

	err := bot.Init()

	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	wg.Wait()
}
