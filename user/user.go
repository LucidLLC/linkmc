package user

import (
	"encoding/json"
	bolt "go.etcd.io/bbolt"
)

type User struct {
	UserID string
	Links  []Link
}

type Link struct {
	Service  string
	Username string
	AddedAt  int64
}

func (u *User) AddLink(link Link) {
	for _, l := range u.Links {
		if l.Service == link.Service {
			return
		}
	}

	u.Links = append(u.Links, link)
}

func (u *User) Save(db *bolt.DB) error {
	return db.Update(func(tx *bolt.Tx) error {
		usersBucket, err := tx.CreateBucketIfNotExists([]byte("users"))

		if err != nil {
			return err
		}

		encoded, err := json.Marshal(u)

		if err != nil {
			return err
		}

		return usersBucket.Put([]byte(u.UserID), encoded)
	})
}

func GetOrCreateUser(uuid string, db *bolt.DB) *User {

}
