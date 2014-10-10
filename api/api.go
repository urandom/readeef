package api

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/urandom/readeef"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
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

	if config.Hubbub.CallbackURL != "" {
		hubbub := readeef.NewHubbub(db, config, logger, dispatcher.Pattern, fm.RemoveFeedChannel(), fm.AddFeedChannel(), updateFeed)
		if err := hubbub.InitSubscriptions(); err != nil {
			return errors.New(fmt.Sprintf("Error initializing hubbub subscriptions: %v", err))
		}

		fm.SetHubbub(hubbub)
		dispatcher.Handle(readeef.NewHubbubController(hubbub))
	}

	var si readeef.SearchIndex
	if config.SearchIndex.BlevePath != "" {
		var err error
		si, err = readeef.NewSearchIndex(config, db, logger)
		if err != nil {
			return errors.New(fmt.Sprintf("Error initializing bleve search: %v", err))
		}

		if si.IsNewIndex() {
			go func() {
				si.IndexAllArticles()
			}()
		}

		fm.SetSearchIndex(si)
	}

	fm.Start()

	nonce := readeef.NewNonce()

	var controller webfw.Controller

	controller = NewAuth()
	dispatcher.Handle(controller)

	controller = NewFeed(fm)
	dispatcher.Handle(controller)

	controller = NewArticle(config)
	dispatcher.Handle(controller)

	if config.SearchIndex.BlevePath != "" {
		controller = NewSearch(si)
		dispatcher.Handle(controller)
	}

	controller = NewUser()
	dispatcher.Handle(controller)

	controller = NewUserSettings()
	dispatcher.Handle(controller)

	controller = NewFeedUpdateNotificator(updateFeed)
	dispatcher.Handle(controller)

	controller = NewNonce(nonce)
	dispatcher.Handle(controller)

	if config.API.Fever {
		controller = NewFever(fm)
		dispatcher.Handle(controller)
	}

	middleware.InitializeDefault(dispatcher)
	dispatcher.RegisterMiddleware(readeef.Auth{DB: db, Pattern: dispatcher.Pattern, Nonce: nonce, IgnoreURLPrefix: config.Auth.IgnoreURLPrefix})

	dispatcher.Context.SetGlobal(readeef.CtxKey("config"), config)
	dispatcher.Context.SetGlobal(context.BaseCtxKey("readeefConfig"), config)
	dispatcher.Context.SetGlobal(readeef.CtxKey("db"), db)

	dispatcher.Renderer = renderer.NewRenderer(dispatcher.Config.Renderer.Dir,
		dispatcher.Config.Renderer.Base)

	dispatcher.Renderer.Delims("{%", "%}")

	go func() {
		for {
			select {
			case <-time.After(5 * time.Minute):
				nonce.Clean(45 * time.Second)
			}
		}
	}()

	return nil
}
