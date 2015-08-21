package api

import (
	"fmt"
	"time"

	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/content/repo"
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

func RegisterControllers(config readeef.Config, dispatcher *webfw.Dispatcher, logger webfw.Logger) error {
	repo, err := repo.New(config.DB.Driver, config.DB.Connect, logger)
	if err != nil {
		return err
	}

	if err := initAdminUser(repo, []byte(config.Auth.Secret)); err != nil {
		return err
	}

	dispatcher.Context.SetGlobal(readeef.CtxKey("config"), config)
	dispatcher.Context.SetGlobal(context.BaseCtxKey("readeefConfig"), config)
	dispatcher.Context.SetGlobal(readeef.CtxKey("repo"), repo)

	um := &readeef.UpdateFeedReceiverManager{}

	fm := readeef.NewFeedManager(repo, config, logger, um)

	if config.Hubbub.CallbackURL != "" {
		hubbub := readeef.NewHubbub(repo, config, logger, dispatcher.Pattern, fm.RemoveFeedChannel(), fm.AddFeedChannel(), um)
		if err := hubbub.InitSubscriptions(); err != nil {
			return fmt.Errorf("Error initializing hubbub subscriptions: %v", err)
		}

		fm.SetHubbub(hubbub)
		dispatcher.Handle(readeef.NewHubbubController(hubbub))
	}

	var si readeef.SearchIndex
	if config.SearchIndex.BlevePath != "" {
		var err error
		si, err = readeef.NewSearchIndex(repo, config, logger)
		if err != nil {
			return fmt.Errorf("Error initializing bleve search: %v", err)
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

	multiPatternController = NewFeed(fm, si)
	dispatcher.Handle(multiPatternController)

	multiPatternController = NewArticle(config)
	dispatcher.Handle(multiPatternController)

	multiPatternController = NewUser()
	dispatcher.Handle(multiPatternController)

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
	dispatcher.RegisterMiddleware(readeef.Auth{Pattern: dispatcher.Pattern, Nonce: nonce, IgnoreURLPrefix: config.Auth.IgnoreURLPrefix})

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

func initAdminUser(repo content.Repo, secret []byte) error {
	users := repo.AllUsers()
	if repo.HasErr() {
		return repo.Err()
	}

	if len(users) > 0 {
		return nil
	}

	u := repo.User()
	u.Data(data.User{Login: data.Login("admin"), Active: true, Admin: true})
	u.Password("admin", secret)
	u.Update()

	return u.Err()
}
