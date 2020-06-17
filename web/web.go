package web

import (
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rbrick/linkmc/config"
	"github.com/rbrick/linkmc/user"
	"github.com/rbrick/linkmc/util"
	bolt "go.etcd.io/bbolt"
	"log"
	"net/http"
	"strconv"
	"time"
)

var (
	messageChannel = make(chan interface{})
)

func QueueMessage(s interface{}) {
	messageChannel <- s
}

type Handler struct {
	db   *bolt.DB
	conf config.Web
	echo *echo.Echo

	upgrader *websocket.Upgrader

	clients map[*websocket.Conn]bool

	server *http.Server
}

func (h *Handler) AddClient(conn *websocket.Conn) {
	h.clients[conn] = true
}

func (h *Handler) Start(addr string) {
	go func() {
		if err := h.echo.Start(addr); err != nil {
			log.Panicln(err)
		}
	}()

	go func() {
		for {
			for k, _ := range h.clients {
				err := k.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second))

				if err != nil {
					delete(h.clients, k)
				} else {
					_, _, _ = k.ReadMessage()
				}
			}
		}
	}()

	go func() {
		for msg := range messageChannel {
			for k, _ := range h.clients {
				err := k.WriteJSON(msg)

				if err != nil {
					delete(h.clients, k)
				}
			}
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

	if err = u.AddPendingLink(pendingLink); err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}

	_ = tx.Rollback()

	tx, err = h.db.Begin(true)

	if err != nil {
		return err
	}

	if err = u.Save(tx); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		_ = tx.Rollback()
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

func (h *Handler) WebsocketHandler(ctx echo.Context) error {
	//go func() {
	wsConn, err := h.upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)

	if err != nil {
		ctx.Logger().Error(err)
		return err
	}
	h.AddClient(wsConn)
	return nil
}

func New(db *bolt.DB, web config.Web) *Handler {
	h := &Handler{
		db:       db,
		echo:     echo.New(),
		conf:     web,
		upgrader: &websocket.Upgrader{},
		clients:  make(map[*websocket.Conn]bool),
	}

	logConfig := middleware.DefaultLoggerConfig
	logConfig.Format = "${time_rfc3339} - ${id} » ${method} ${host}${path} (${status})\n"

	h.echo.Use(h.AuthTokenMiddleware, middleware.LoggerWithConfig(logConfig))
	h.echo.GET("/links/:uuid", h.getLinks)
	h.echo.GET("/startlink/:service/:uuid/:expire", h.startLink)
	h.echo.GET("/messages", h.WebsocketHandler)

	return h
}
