package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rbrick/linkmc/bot"
	"github.com/rbrick/linkmc/user"
	bolt "go.etcd.io/bbolt"
	"time"
)

func verify(bot bot.Context, name string, args []string) {
	if len(args) != 1 {
		bot.SendMessage("Invalid number of arguments")
		return
	}

	if args[0] == "" {
		bot.SendMessage("Invalid argument")
		return
	}

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

		if bot.Bot().Config().Name != pendingLink.Service {
			return errors.New(fmt.Sprintf("mismatched service. Token for %s used on %s.", pendingLink.Service, bot.Bot().Config().Name))
		}

		if time.Now().Unix() >= pendingLink.Expire {
			return user.ErrPendingLinkExpired
		}

		u, err := user.GetOrCreateUser(pendingLink.UserID, tx)

		if err != nil {
			return err
		}

		added := u.AddLink(user.Link{
			Service:  bot.Bot().Config().Name,
			Username: bot.User(),
			AddedAt:  time.Now().Unix(),
		})

		if !added {
			return errors.New("service already linked")
		}

		u.RemovePendingLink(pendingLink.Service)

		_ = keys.Delete([]byte(args[0]))

		return u.Save(tx)
	})

	if err != nil {
		bot.SendMessage(err.Error())
	} else {
		bot.SendMessage("You have successfully linked your account!")
	}

}
