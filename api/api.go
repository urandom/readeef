package api

import (
	"fmt"
	"time"

	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base/extractor"
	"github.com/urandom/readeef/content/base/monitor"
	contentProcessor "github.com/urandom/readeef/content/base/processor"
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

	capabilities := capabilities{
		I18N:       len(dispatcher.Config.I18n.Languages) > 1,
		Popularity: len(config.Popularity.Providers) > 0,
	}

	var ap []content.ArticleProcessor
	for _, p := range config.Content.ArticleProcessors {
		switch p {
		case "relative-url":
			ap = append(ap, contentProcessor.NewRelativeUrl(logger))
		case "proxy-http":
			template := config.Content.ProxyHTTPURLTemplate

			if template != "" {
				p, err := contentProcessor.NewProxyHTTP(logger, template)
				if err != nil {
					return fmt.Errorf("Error initializing Proxy HTTP article processor: %v", err)
				}
				ap = append(ap, p)
				capabilities.ProxyHTTP = true
			}
		case "insert-thumbnail-target":
			ap = append(ap, contentProcessor.NewInsertThumbnailTarget(logger))
		}
	}

	repo.ArticleProcessors(ap)

	if err := initAdminUser(repo, []byte(config.Auth.Secret)); err != nil {
		return err
	}

	mw := make([]string, 0, len(dispatcher.Config.Dispatcher.Middleware))
	for _, m := range dispatcher.Config.Dispatcher.Middleware {
		switch m {
		case "I18N", "Static", "Url", "Sitemap":
		case "Session":
			if capabilities.ProxyHTTP {
				mw = append(mw, m)
			}
		default:
			mw = append(mw, m)
		}
	}

	dispatcher.Config.Dispatcher.Middleware = mw

	dispatcher.Context.SetGlobal(readeef.CtxKey("config"), config)
	dispatcher.Context.SetGlobal(context.BaseCtxKey("readeefConfig"), config)
	dispatcher.Context.SetGlobal(readeef.CtxKey("repo"), repo)

	fm := readeef.NewFeedManager(repo, config, logger)

	var processors []parser.Processor
	for _, p := range config.FeedParser.Processors {
		switch p {
		case "absolutize-urls":
			processors = append(processors, processor.NewAbsolutizeURLs(logger))
		case "relative-url":
			processors = append(processors, processor.NewRelativeUrl(logger))
		case "proxy-http":
			template := config.FeedParser.ProxyHTTPURLTemplate

			if template != "" {
				p, err := processor.NewProxyHTTP(logger, template)
				if err != nil {
					return fmt.Errorf("Error initializing Proxy HTTP processor: %v", err)
				}
				processors = append(processors, p)
				capabilities.ProxyHTTP = true
			}
		case "cleanup":
			processors = append(processors, processor.NewCleanup(logger))
		case "top-image-marker":
			processors = append(processors, processor.NewTopImageMarker(logger))
		}
	}

	fm.ParserProcessors(processors)

	var sp content.SearchProvider

	switch config.Content.SearchProvider {
	case "elastic":
		if sp, err = search.NewElastic(config.Content.ElasticURL, config.Content.SearchBatchSize, logger); err != nil {
			logger.Printf("Error initializing Elastic search: %v\n", err)
		}
	case "bleve":
		fallthrough
	default:
		if sp, err = search.NewBleve(config.Content.BlevePath, config.Content.SearchBatchSize, logger); err != nil {
			logger.Printf("Error initializing Bleve search: %v\n", err)
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

	if ce != nil {
		capabilities.Extractor = true
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

	monitors := []content.FeedMonitor{monitor.NewUnread(repo, logger)}
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
		hubbub := readeef.NewHubbub(repo, config, logger, dispatcher.Pattern,
			fm.RemoveFeedChannel())
		if err := hubbub.InitSubscriptions(); err != nil {
			return fmt.Errorf("Error initializing hubbub subscriptions: %v", err)
		}

		hubbub.FeedMonitors(monitors)
		fm.Hubbub(hubbub)
	}

	fm.FeedMonitors(monitors)

	fm.Start()

	nonce := readeef.NewNonce()

	controllers := []webfw.Controller{
		NewAuth(capabilities),
		NewFeed(fm, sp),
		NewArticle(config, ce),
		NewUser(),
		NewUserSettings(),
		NewNonce(nonce),
	}

	if fm.Hubbub() != nil {
		controllers = append(controllers, NewHubbubController(fm.Hubbub(), config.Hubbub.RelativePath,
			fm.AddFeedChannel(), fm.RemoveFeedChannel()))
	}

	for _, e := range config.API.Emulators {
		switch e {
		case "tt-rss":
			controllers = append(controllers, NewTtRss(fm, sp))
		case "fever":
			controllers = append(controllers, NewFever())
		}
	}

	for _, c := range controllers {
		dispatcher.Handle(c)
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
