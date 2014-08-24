package api

import (
	"errors"
	"fmt"
	"log"
	"readeef"

	"github.com/urandom/webfw"
	"github.com/urandom/webfw/middleware"
)

func RegisterControllers(config readeef.Config, dispatcher *webfw.Dispatcher, logger *log.Logger) error {
	db := readeef.NewDB(config.DB.Driver, config.DB.Connect)
	if err := db.Connect(); err != nil {
		return errors.New(fmt.Sprintf("Error connecting to database: %v", err))
	}

	updateFeed := make(chan readeef.Feed)

	fm := readeef.NewFeedManager(db, config, logger, updateFeed)

	if config.Hubbub.CallbackURL != "" {
		hubbub := readeef.NewHubbub(db, config, logger, fm.RemoveFeedChannel(), fm.AddFeedChannel(), updateFeed)

		dispatcher.Handle(readeef.NewHubbubController(hubbub))
	}

	var controller webfw.Controller

	controller = NewAuth()
	dispatcher.Handle(controller)

	controller = NewFeed(fm)
	dispatcher.Handle(controller)

	middleware.InitializeDefault(dispatcher)
	dispatcher.RegisterMiddleware(readeef.Auth{DB: db, Pattern: dispatcher.Pattern})

	dispatcher.Context.SetGlobal(readeef.CtxKey("config"), config)
	dispatcher.Context.SetGlobal(readeef.CtxKey("db"), db)

	return nil
}
