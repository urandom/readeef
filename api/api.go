package api

import (
	"errors"
	"fmt"
	"readeef"

	"github.com/urandom/webfw"
)

func RegisterControllers(config readeef.Config, dispatcher *webfw.Dispatcher) error {
	db := readeef.NewDB(config.DB.Driver, config.DB.Connect)
	if err := db.Connect(); err != nil {
		return errors.New(fmt.Sprintf("Error connecting to database: %v", err))
	}

	fm := readeef.NewFeedManager(db, config)
	var controller webfw.Controller

	controller = NewAuth()
	dispatcher.Handle(controller)

	controller = NewFeed(fm)
	dispatcher.Handle(controller)

	if config.Hubbub.CallbackURL != "" {
		hubbub := readeef.NewHubbub(db, config)

		dispatcher.Handle(readeef.NewHubbubController(hubbub))
	}

	dispatcher.RegisterMiddleware(readeef.Auth{DB: db, Pattern: dispatcher.Pattern})

	return nil
}
