package user

import (
	"encoding/json"
	"errors"
	"fmt"
	bolt "go.etcd.io/bbolt"
	"time"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrLinkAlreadyPending = errors.New("link already pending")
	ErrPendingLinkExpired = errors.New("link expired")
	ErrAlreadyLinked      = errors.New("account already linked")
)

type User struct {
	UserID       string        `json:"id"`
	Links        []Link        `json:"links"`
	PendingLinks []PendingLink `json:"pending_links"`
}

type Link struct {
	Service  string `json:"service"`
	Username string `json:"username"`
	AddedAt  int64  `json:"addedAt"`
}

type PendingLink struct {
	Service string `json:"service"`
	UserID  string `json:"id"`
	Expire  int64  `json:"expire"`
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

func (u *User) AddPendingLink(link PendingLink) error {

	for _, l := range u.Links {
		if l.Service == link.Service {
			return ErrAlreadyLinked
		}
	}

	for _, l := range u.PendingLinks {
		if l.Service == link.Service {
			now := time.Now().UnixNano() / 1000
			if u.UserID == "2e0fe8c0-5b79-42b8-97a6-8db61a374983" {
				fmt.Println("current time:", now, " expire time:", link.Expire)
			}

			if now <= link.Expire {
				return ErrLinkAlreadyPending
			}
		}
	}

	u.RemovePendingLink(link.Service)
	u.PendingLinks = append(u.PendingLinks, link)
	return nil
}

func (u *User) RemovePendingLink(service string) {
	idx := -1
	for i, l := range u.PendingLinks {
		if l.Service == service {
			idx = i
		}
	}

	if idx != -1 {
		x := u.PendingLinks[len(u.PendingLinks)-1]
		u.PendingLinks[len(u.PendingLinks)-1] = PendingLink{}
		u.PendingLinks[idx] = x

		u.PendingLinks = u.PendingLinks[:len(u.PendingLinks)-1]
	}
}

func (u *User) Save(tx *bolt.Tx) error {

	bucket, err := tx.CreateBucketIfNotExists([]byte("users"))

	if err != nil {
		return err
	}

	marshaled, err := json.Marshal(u)

	if err != nil {
		return err
	}

	return bucket.Put([]byte(u.UserID), marshaled)
}

func NewUser(uuid string) *User {
	return &User{
		UserID: uuid,
		Links:  []Link{},
	}
}

func GetUser(uuid string, db *bolt.DB) (*User, error) {
	tx, err := db.Begin(false)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	bucket := tx.Bucket([]byte("users"))

	if bucket == nil {
		return nil, errors.New("bucket not found")
	}

	userJson := bucket.Get([]byte(uuid))

	var user User

	if userJson != nil {
		if err := json.Unmarshal(userJson, &user); err != nil {
			return nil, err
		}

		return &user, nil
	}

	return nil, ErrUserNotFound
}

func GetOrCreateUser(uuid string, tx *bolt.Tx) (*User, error) {
	//tx, err := db.Begin(true)
	//
	//if err != nil {
	//	return nil, err
	//}
	//
	//defer tx.Rollback()
	bucket, err := tx.CreateBucketIfNotExists([]byte("users"))

	if err != nil {
		return nil, err
	}

	userJson := bucket.Get([]byte(uuid))

	var user User

	if userJson != nil {
		if err = json.Unmarshal(userJson, &user); err != nil {
			return nil, err
		}
	} else {
		user = *NewUser(uuid)
		d, err := json.Marshal(user)

		if err != nil {
			return nil, err
		}

		_ = bucket.Put([]byte(uuid), d)
	}

	return &user, nil
}
