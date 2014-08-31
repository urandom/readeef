package api

import (
	"errors"
	"fmt"
	"log"
	"readeef"
	"time"

	"github.com/urandom/webfw"
	"github.com/urandom/webfw/middleware"
	"github.com/urandom/webfw/renderer"
)

func RegisterControllers(config readeef.Config, dispatcher *webfw.Dispatcher, logger *log.Logger) error {
	db := readeef.NewDB(config.DB.Driver, config.DB.Connect)
	if err := db.Connect(); err != nil {
		return errors.New(fmt.Sprintf("Error connecting to database: %v", err))
	}

	updateFeed := make(chan readeef.Feed)

	fm := readeef.NewFeedManager(db, config, logger, updateFeed)
	fm.Start()

	if config.Hubbub.CallbackURL != "" {
		hubbub := readeef.NewHubbub(db, config, logger, fm.RemoveFeedChannel(), fm.AddFeedChannel(), updateFeed)

		dispatcher.Handle(readeef.NewHubbubController(hubbub))
	}

	nonce := readeef.NewNonce()

	var controller webfw.Controller

	controller = NewAuth()
	dispatcher.Handle(controller)

	controller = NewFeed(fm)
	dispatcher.Handle(controller)

	controller = NewArticle()
	dispatcher.Handle(controller)

	controller = NewNonce(nonce)
	dispatcher.Handle(controller)

	middleware.InitializeDefault(dispatcher)
	dispatcher.RegisterMiddleware(readeef.Auth{DB: db, Pattern: dispatcher.Pattern, Nonce: nonce})

	dispatcher.Context.SetGlobal(readeef.CtxKey("config"), config)
	dispatcher.Context.SetGlobal(readeef.CtxKey("db"), db)

	dispatcher.Renderer = renderer.NewRenderer(dispatcher.Config.Renderer.Dir,
		dispatcher.Config.Renderer.Base)

	dispatcher.Renderer.Delims("{%", "%}")

	go func() {
		for {
			select {
			case f := <-updateFeed:
				readeef.Debug.Println("Feed " + f.Link + " updated")
			case <-time.After(5 * time.Minute):
				nonce.Clean(45 * time.Second)
			}
		}
	}()

	return nil
}
