package web

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/rbrick/linkmc/config"
	"github.com/rbrick/linkmc/user"
	"github.com/rbrick/linkmc/util"
	bolt "go.etcd.io/bbolt"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Handler struct {
	db   *bolt.DB
	conf config.Web
	echo *echo.Echo

	server *http.Server
}

func (h *Handler) Start(addr string) {
	go func() {
		if err := h.echo.Start(addr); err != nil {
			log.Panicln(err)
		}
	}()

	log.Printf("started http server on %s\n", addr)
}

func (h *Handler) Shutdown() error {
	log.Println("shutting down http server...")
	return h.echo.Shutdown(context.Background())
}

func (h *Handler) getLinks(ctx echo.Context) error {
	key := ctx.Param("uuid")

	u, err := user.GetUser(key, h.db)

	if err != nil {
		if err == user.ErrUserNotFound {
			return ctx.String(http.StatusNotFound, err.Error())
		}
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(200, u)
}

func (h *Handler) startLink(ctx echo.Context) error {

	service := ctx.Param("service")
	uuid := ctx.Param("uuid")
	expireStr := ctx.Param("expire")

	expire, _ := strconv.ParseInt(expireStr, 10, 0)

	// Create a pending link
	pendingLink := user.PendingLink{
		Service: service,
		UserID:  uuid,
		Expire:  time.Now().Add(time.Duration(expire) * time.Minute).Unix(),
	}

	tx, err := h.db.Begin(true)

	if err != nil {
		return err
	}

	// Add the pending link to the user if possible
	u, err := user.GetOrCreateUser(uuid, tx)
	if err != nil {
		// link is already pending
		if err == user.ErrLinkAlreadyPending {
			return ctx.String(http.StatusBadRequest, err.Error())
		}

		// an unknown err occurred
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	if err = tx.Commit(); err != nil {
		_ = tx.Rollback()
		return err
	}

	if !u.AddPendingLink(pendingLink) {

		return ctx.String(http.StatusBadRequest, user.ErrLinkAlreadyPending.Error())
	}

	_ = tx.Rollback()

	tx, err = h.db.Begin(true)

	if err != nil {
		fmt.Println("error saving user")
		return err
	}

	if err = u.Save(tx); err != nil {
		fmt.Println("error saving user")
		return err
	}

	if err = tx.Commit(); err != nil {
		_ = tx.Rollback()
		fmt.Println("error saving user")
		return err
	}

	// Try to update the keys
	// Random String -> Pending link
	err = h.db.Update(func(tx *bolt.Tx) error {
		keys, err := tx.CreateBucketIfNotExists([]byte("keys"))

		if err != nil {
			return err
		}

		// ensure the key is random
		rndStr := util.RandomString(time.Now().UnixNano(), 8)

		for keys.Get([]byte(rndStr)) != nil {
			rndStr = util.RandomString(time.Now().UnixNano(), 8)
		}

		// encode the link into json
		b, err := json.Marshal(&pendingLink)

		// error occurred marshalling the string
		if err != nil {
			return err
		}

		// try to insert the key into the database
		if err = keys.Put([]byte(rndStr), b); err != nil {
			return err // error occurred
		}

		// return the random string for use in messages
		return ctx.String(200, rndStr)
	})

	return err
}

func (h *Handler) AuthTokenMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		if ctx.Request().Header.Get("X-Link-Auth-Token") == h.conf.AuthToken {
			return next(ctx)
		}
		return ctx.String(http.StatusForbidden, "invalid authentication token")
	}
}

func New(db *bolt.DB, web config.Web) *Handler {
	h := &Handler{
		db:   db,
		echo: echo.New(),
		conf: web,
	}

	h.echo.GET("/links/:uuid", h.getLinks, h.AuthTokenMiddleware)
	h.echo.GET("/startlink/:service/:uuid/:expire", h.startLink, h.AuthTokenMiddleware)

	return h
}
