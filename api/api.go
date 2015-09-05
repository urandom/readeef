package api

import (
	"fmt"
	"time"

	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base/extractor"
	"github.com/urandom/readeef/content/base/monitor"
	"github.com/urandom/readeef/content/base/search"
	"github.com/urandom/readeef/content/base/thumbnailer"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/parser"
	"github.com/urandom/readeef/parser/processor"
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

	fm := readeef.NewFeedManager(repo, config, logger)

	capabilities := capabilities{}

	var sp content.SearchProvider

	switch config.Content.SearchProvider {
	case "elastic":
		if sp, err = search.NewElastic(config.Content.ElasticURL, config.Content.SearchBatchSize, logger); err != nil {
			return fmt.Errorf("Error initializing Elastic search: %v\n", err)
		}
	case "bleve":
		fallthrough
	default:
		if sp, err = search.NewBleve(config.Content.BlevePath, config.Content.SearchBatchSize, logger); err != nil {
			return fmt.Errorf("Error initializing Bleve search: %v\n", err)
		}
	}

	if sp != nil {
		if sp.IsNewIndex() {
			go func() {
				sp.IndexAllFeeds(repo)
			}()
		}
	}

	var ce content.Extractor

	switch config.Content.Extractor {
	case "readability":
		if ce, err = extractor.NewReadability(config.Content.ReadabilityKey); err != nil {
			return fmt.Errorf("Error initializing Readability extractor: %v\n", err)
		}
	case "goose":
		fallthrough
	default:
		if ce, err = extractor.NewGoose(dispatcher.Config.Renderer.Dir); err != nil {
			return fmt.Errorf("Error initializing Goose extractor: %v\n", err)
		}
	}

	var t content.Thumbnailer
	switch config.Content.Thumbnailer {
	case "extract":
		if t, err = thumbnailer.NewExtract(ce, logger); err != nil {
			return fmt.Errorf("Error initializing Extract thumbnailer: %v\n", err)
		}
	case "description":
		fallthrough
	default:
		t = thumbnailer.NewDescription(logger)
	}

	var monitors []content.FeedMonitor
	for _, m := range config.FeedManager.Monitors {
		switch m {
		case "index":
			if sp != nil {
				monitors = append(monitors, monitor.NewIndex(sp, logger))
				capabilities.Search = true
			}
		case "thumbnailer":
			if t != nil {
				monitors = append(monitors, monitor.NewThumbnailer(t, logger))
			}
		}
	}

	webSocket := NewWebSocket(fm, sp, ce, capabilities)
	dispatcher.Handle(webSocket)

	monitors = append(monitors, webSocket)

	if config.Hubbub.CallbackURL != "" {
		hubbub := readeef.NewHubbub(repo, config, logger, dispatcher.Pattern, fm.RemoveFeedChannel(), fm.AddFeedChannel())
		if err := hubbub.InitSubscriptions(); err != nil {
			return fmt.Errorf("Error initializing hubbub subscriptions: %v", err)
		}

		hubbub.FeedMonitors(monitors)
		fm.Hubbub(hubbub)
		dispatcher.Handle(readeef.NewHubbubController(hubbub))
	}

	fm.FeedMonitors(monitors)

	var processors []parser.Processor
	for _, p := range config.FeedParser.Processors {
		switch p {
		case "relative-url":
			processors = append(processors, processor.NewRelativeUrl(logger))
		case "cleanup":
			processors = append(processors, processor.NewCleanup(logger))
		}
	}

	fm.ParserProcessors(processors)

	fm.Start()

	nonce := readeef.NewNonce()

	var patternController webfw.PatternController
	var multiPatternController webfw.MultiPatternController

	patternController = NewAuth(capabilities)
	dispatcher.Handle(patternController)

	multiPatternController = NewFeed(fm, sp)
	dispatcher.Handle(multiPatternController)

	multiPatternController = NewArticle(config, ce)
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
