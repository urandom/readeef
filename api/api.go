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

type responseError struct {
	val     map[string]interface{}
	err     error
	errType string
}

func newResponse() responseError {
	return responseError{val: make(map[string]interface{})}
}

func RegisterControllers(config readeef.Config, dispatcher *webfw.Dispatcher, logger *log.Logger) error {
	db := readeef.NewDB(config.DB.Driver, config.DB.Connect)
	if err := db.Connect(); err != nil {
		return errors.New(fmt.Sprintf("Error connecting to database: %v", err))
	}

	um := &readeef.UpdateFeedReceiverManager{}

	fm := readeef.NewFeedManager(db, config, logger, um)

	if config.Hubbub.CallbackURL != "" {
		hubbub := readeef.NewHubbub(db, config, logger, dispatcher.Pattern, fm.RemoveFeedChannel(), fm.AddFeedChannel(), um)
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

	var patternController webfw.PatternController
	var multiPatternController webfw.MultiPatternController

	patternController = NewAuth()
	dispatcher.Handle(patternController)

	multiPatternController = NewFeed(fm)
	dispatcher.HandleMultiPattern(multiPatternController)

	multiPatternController = NewArticle(config)
	dispatcher.HandleMultiPattern(multiPatternController)

	if config.SearchIndex.BlevePath != "" {
		patternController = NewSearch(si)
		dispatcher.Handle(patternController)
	}

	multiPatternController = NewUser()
	dispatcher.HandleMultiPattern(multiPatternController)

	patternController = NewUserSettings()
	dispatcher.Handle(patternController)

	patternController = NewNonce(nonce)
	dispatcher.Handle(patternController)

	if config.API.Fever {
		patternController = NewFever(fm)
		dispatcher.Handle(patternController)
	}

	webSocket := NewWebSocket(fm, si)
	dispatcher.Handle(webSocket)
	um.AddUpdateReceiver(webSocket)

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
