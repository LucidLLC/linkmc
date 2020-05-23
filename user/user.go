package user

import (
	"encoding/json"
	bolt "go.etcd.io/bbolt"
)

type User struct {
	UserID string `json:"id"`
	Links  []Link `json:"links"`
}

type Link struct {
	Service  string `json:"service"`
	Username string `json:"username"`
	AddedAt  int64  `json:"addedAt"`
}

func (u *User) AddLink(link Link) bool {
	for _, l := range u.Links {
		if l.Service == link.Service {
			return false
		}
	}

	u.Links = append(u.Links, link)
	return true
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

func NewUser(uuid string) *User {
	return &User{
		UserID: uuid,
		Links:  []Link{},
	}
}

func GetOrCreateUser(uuid string, db *bolt.DB, cb func(*User) error) error {
	return db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("users"))

		if err != nil {
			return err
		}

		userJson := bucket.Get([]byte(uuid))

		var user *User

		if userJson != nil {
			if err = json.Unmarshal(userJson, user); err != nil {
				return err
			}
		} else {
			user = NewUser(uuid)
		}

		return cb(user)
	})

}
