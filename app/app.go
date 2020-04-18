package app

import (
	"github.com/rbrick/linkmc/config"
	"sync"
)

type Application struct {
	conf *config.Config
	wg   sync.WaitGroup
}

func (app *Application) Run() {

}

func New(conf *config.Config) *Application {
	return &Application{conf: conf}
}
