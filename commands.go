package main

import (
	"errors"
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

	err := application.DB.Update(func(tx *bolt.Tx) error {
		keys, err := tx.CreateBucketIfNotExists([]byte("keys"))

		if err != nil {
			return err
		}

		uuid := keys.Get([]byte(args[0]))

		if uuid == nil {
			return errors.New("invalid key")
		}

		return user.GetOrCreateUser(string(uuid), application.DB, func(u *user.User) error {
			added := u.AddLink(user.Link{
				Service:  bot.Bot().Config().Name,
				Username: bot.User(),
				AddedAt:  time.Now().Unix(),
			})

			if !added {
				return errors.New("service already linked")
			}

			return nil
		})
	})

	if err != nil {
		bot.SendMessage(err.Error())
	} else {
		bot.SendMessage("You have successfully linked your account!")
	}

}
