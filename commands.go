package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/LucidLLC/linkmc/bot"
	"github.com/LucidLLC/linkmc/user"
	"github.com/LucidLLC/linkmc/web"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	//tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	bolt "go.etcd.io/bbolt"
	"time"
)

func verify(context bot.Context, name string, args []string) {
	if len(args) != 1 {
		context.SendMessage("Invalid number of arguments")
		return
	}

	if args[0] == "" {
		context.SendMessage("Invalid argument")
		return
	}

	if context.Bot().Config().Name == "telegram" {
		tgapi := context.Bot().API().(*tgbotapi.BotAPI)
		channel := context.Bot().(*bot.TelegramBot).BroadcastChannel

		member, err := tgapi.GetChatMember(tgbotapi.ChatConfigWithUser{
			UserID:             context.UserID(),
			SuperGroupUsername: channel,
		})

		if err != nil {
			context.SendMessage(err.Error())
			return
		} else if !member.IsMember() {
			context.SendMessage("You do not appear to be in our news channel! Please join @MCTeamsNews!")
			return
		}
	}

	var fUser *user.User
	var fLink user.Link
	err := application.DB.Update(func(tx *bolt.Tx) error {
		keys, err := tx.CreateBucketIfNotExists([]byte("keys"))

		if err != nil {
			return err
		}

		jsonLink := keys.Get([]byte(args[0]))

		if jsonLink == nil {
			return errors.New("invalid key")
		}

		var pendingLink user.PendingLink

		if err = json.Unmarshal(jsonLink, &pendingLink); err != nil {
			return err
		}

		if context.Bot().Config().Name != pendingLink.Service {
			return errors.New(fmt.Sprintf("mismatched service. Token for %s used on %s.", pendingLink.Service, context.Bot().Config().Name))
		}

		if time.Now().Unix() >= pendingLink.Expire {
			return user.ErrPendingLinkExpired
		}

		u, err := user.GetOrCreateUser(pendingLink.UserID, tx)

		if err != nil {
			return err
		}

		fLink = user.Link{
			Service:  context.Bot().Config().Name,
			Username: context.User(),
			AddedAt:  time.Now().Unix(),
		}
		added := u.AddLink(fLink)

		if !added {
			return errors.New("service already linked")
		}

		u.RemovePendingLink(pendingLink.Service)

		_ = keys.Delete([]byte(args[0]))

		fUser = u
		return u.Save(tx)
	})

	if err != nil {
		context.SendMessage(err.Error())
	} else {
		web.QueueMessage(struct {
			U    *user.User `json:"user"`
			Link user.Link  `json:"link"`
		}{
			U:    fUser,
			Link: fLink,
		})
		context.SendMessage("You have successfully linked your account!")
	}

}
