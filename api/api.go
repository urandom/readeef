package api

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	"github.com/urandom/handler/auth"
	"github.com/urandom/handler/method"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base/extractor"
	contentProcessor "github.com/urandom/readeef/content/base/processor"
	"github.com/urandom/readeef/content/base/search"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/parser"
	"github.com/urandom/readeef/parser/processor"
)

/*
func RegisterControllers(config readeef.Config, dispatcher *webfw.Dispatcher, logger webfw.Logger) error {
	repo, err := repo.New(config.DB.Driver, config.DB.Connect, logger)
	if err != nil {
		return err
	}

	capabilities := capabilities{
		I18N:       len(config.I18n.Languages) > 1,
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
*/

func Prepare(config readeef.Config, log readeef.Logger) error {
	repo, err := repo.New(config.DB.Driver, config.DB.Connect, log)
	if err != nil {
		return errors.Wrap(err, "creating repo")
	}

	features := features{
		I18N:       len(config.I18n.Languages) > 1,
		Popularity: len(config.Popularity.Providers) > 0,
	}

	if processors, err := initArticleProcessors(config.Content.ArticleProcessors, config.Content.ProxyHTTPURLTemplate, &features, log); err == nil {
		repo.ArticleProcessors(processors)
	} else {
		return errors.Wrap(err, "initializing article processors")
	}

	if err := initAdminUser(repo, []byte(config.Auth.Secret)); err != nil {
		return errors.Wrap(err, "initializing admin user")
	}

	feedManager := readeef.NewFeedManager(repo, config, log)

	if processors, err := initParserProcessors(config.FeedParser.Processors, config.FeedParser.ProxyHTTPURLTemplate, &features, log); err == nil {
		feedManager.ParserProcessors(processors)
	} else {
		return errors.Wrap(err, "initializing parser processors")
	}

	searchProvider := initSearchProvider(config, repo, log)

	if searchProvider != nil {
		features.Search = true
	}

	extractor, err := initContentExtractor(config)
	if err != nil {
		return errors.Wrap(err, "initializing content extractor")
	}

	if extractor != nil {
		features.Extractor = true
	}

	return nil
}

func initArticleProcessors(names []string, proxyTemplate string, features *features, log readeef.Logger) ([]content.ArticleProcessor, error) {
	var processors []content.ArticleProcessor

	for _, p := range names {
		switch p {
		case "relative-url":
			processors = append(processors, contentProcessor.NewRelativeURL(log))
		case "proxy-http":
			template := proxyTemplate

			if template != "" {
				p, err := contentProcessor.NewProxyHTTP(template, log)
				if err != nil {
					return nil, errors.Wrap(err, "initializing proxy http processor")
				}
				processors = append(processors, p)
				features.ProxyHTTP = true
			}
		case "insert-thumbnail-target":
			processors = append(processors, contentProcessor.NewInsertThumbnailTarget(log))
		}
	}

	return processors, nil
}

func initParserProcessors(names []string, proxyTemplate string, features *features, log readeef.Logger) ([]parser.Processor, error) {
	var processors []parser.Processor

	for _, p := range names {
		switch p {
		case "absolutize-urls":
			processors = append(processors, processor.NewAbsolutizeURLs(log))
		case "relative-url":
			processors = append(processors, processor.NewRelativeURL(log))
		case "proxy-http":
			template := proxyTemplate

			if template != "" {
				p, err := processor.NewProxyHTTP(template, log)
				if err != nil {
					return nil, errors.Wrap(err, "initializing proxy http processor")
				}
				processors = append(processors, p)
				features.ProxyHTTP = true
			}
		case "cleanup":
			processors = append(processors, processor.NewCleanup(log))
		case "top-image-marker":
			processors = append(processors, processor.NewTopImageMarker(log))
		}
	}

	return processors, nil
}

func initSearchProvider(config readeef.Config, repo content.Repo, log readeef.Logger) content.SearchProvider {
	var searchProvider content.SearchProvider
	var err error

	switch config.Content.SearchProvider {
	case "elastic":
		if searchProvider, err = search.NewElastic(
			config.Content.ElasticURL,
			config.Content.SearchBatchSize,
			log,
		); err != nil {
			log.Printf("Error initializing Elastic search: %v\n", err)
		}
	case "bleve":
		fallthrough
	default:
		if searchProvider, err = search.NewBleve(
			config.Content.BlevePath,
			config.Content.SearchBatchSize,
			log,
		); err != nil {
			log.Printf("Error initializing Bleve search: %v\n", err)
		}
	}

	if searchProvider != nil {
		if searchProvider.IsNewIndex() {
			go func() {
				searchProvider.IndexAllFeeds(repo)
			}()
		}
	}

	return searchProvider
}

func initContentExtractor(config readeef.Config) (content.Extractor, error) {
	switch config.Content.Extractor {
	case "readability":
		if ce, err := extractor.NewReadability(config.Content.ReadabilityKey); err == nil {
			return ce, nil
		} else {
			return nil, errors.Wrap(err, "initializing Readability extractor")
		}
	case "goose":
		fallthrough
	default:
		//TODO: pass the filesystem to the goose extractor
		if ce, err := extractor.NewGoose("templates"); err == nil {
			return ce, nil
		} else {
			return nil, errors.Wrap(err, "initializing Goose extractor")
		}
	}
}

func Mux(
	ctx context.Context,
	repo content.Repo,
	config readeef.Config,
	storage content.TokenStorage,
	extractor content.Extractor,
	searchProvider content.SearchProvider,
	feedManager *readeef.FeedManager,
	hubbub *readeef.Hubbub,
	features features,
	log readeef.Logger,
) (http.Handler, error) {
	secret := []byte(config.Auth.Secret)

	r := chi.NewRouter()

	r.Route("/api", func(r chi.Router) {
		r.Route("/v2/token", func(r chi.Router) {
			r.Method(method.POST, "/", tokenCreate(repo, secret, log))
			r.Method(method.DELETE, "/", tokenDelete(storage, secret, log))
		})

		r.Route("/v2/hubbub", func(r chi.Router) {
			r.Get("/", hubbubRegistration(hubbub, repo, feedManager.AddFeedChannel(), feedManager.RemoveFeedChannel(), log))
			r.Post("/", hubbubRegistration(hubbub, repo, feedManager.AddFeedChannel(), feedManager.RemoveFeedChannel(), log))
		})

		r.Route("/v2", func(r chi.Router) {
			r.Use(func(next http.Handler) http.Handler {
				return auth.RequireToken(next, tokenValidator(repo, storage, log), secret)
			})
			r.Use(func(next http.Handler) http.Handler {
				return userContext(repo, next, log)
			})
			r.Use(userValidator)

			r.Get("/features", featuresHandler(features))

			r.Route("/feed", func(r chi.Router) {
				r.Get("/", listFeeds)
				r.Post("/", addFeed(feedManager))

				r.Get("/discover", discoverFeeds(feedManager))

				r.Route("/{feedId:[0-9]+}", func(r chi.Router) {
					r.Use(feedContext)

					r.Delete("/", deleteFeed(feedManager))

					r.Get("/tags", getFeedTags)
					r.Post("/tags", setFeedTags)

				})
			})

			r.Route("/article", func(r chi.Router) {
				r.Route("/{articleId:[0-9]+}", func(r chi.Router) {
					r.Use(articleContext)

					r.Get("/", getArticle)
					if extractor != nil {
						r.Get("/format", formatArticle(extractor))
					}
					r.Post("/read", articleStateChange(read))
					r.Post("/favorite", articleStateChange(favorite))
				})

				r.Route("/favorite", func(r chi.Router) {
					r.Get("/", getArticles(favoriteRepoType))

					r.Post("/read", articlesReadStateChange(favoriteRepoType))
				})

				r.Route("/popular", func(r chi.Router) {
					r.With(feedContext).Get("/feed/{feedId:[0-9]+}",
						getArticles(popularRepoType, feedRepoType))
					r.With(tagContext).Get("/tag/{tagId:[0-9]+}",
						getArticles(popularRepoType, tagRepoType))
					r.Get("/", getArticles(popularRepoType, userRepoType))
				})

				r.Route("/feed/{feedId:[0-9]+}", func(r chi.Router) {
					r.Use(feedContext)

					r.Get("/", getArticles(feedRepoType))

					r.Post("/read", articlesReadStateChange(feedRepoType))
				})

				r.Route("/tag/{tagId:[0-9]+}", func(r chi.Router) {
					r.Use(tagContext)

					r.Get("/", getArticles(tagRepoType))

					r.Post("/read", articlesReadStateChange(tagRepoType))
				})

				r.Get("/", getArticles(userRepoType))

				r.Route("/search", func(r chi.Router) {
					r.Get("/*", articleSearch(searchProvider, userRepoType))
					r.With(feedContext).Get("/feed/{feedId:[0-9]+}/*", articleSearch(searchProvider, feedRepoType))
					r.With(tagContext).Get("/tag/{tagId:[0-9]+}/*", articleSearch(searchProvider, tagRepoType))
				})

				r.Post("/read", articlesReadStateChange(userRepoType))
			})

			r.Route("/opml", func(r chi.Router) {
				r.Get("/", exportOPML(feedManager))
				r.Post("/", importOPML(feedManager))
			})

			r.Get("/events", eventSocket(ctx, storage, feedManager))

			r.Route("/user", func(r chi.Router) {
				r.Route("/", func(r chi.Router) {
					r.Use(adminValidator)

					r.Get("/", listUsers)

					r.Post("/", addUser(secret))
					r.Delete("/{name}", deleteUser)

					r.Post("/{name}/settings/{key}", setSettingValue(secret))
				})

				r.Get("/data", getUserData)

				r.Route("/settings", func(r chi.Router) {
					r.Get("/", getSettingKeys)
					r.Get("/{key}", getSettingValue)
					r.Post("/{key}", setSettingValue(secret))
				})
			})
		})
	})

	return r, nil
}

func tokenValidator(
	repo content.Repo,
	storage content.TokenStorage,
	log readeef.Logger,
) auth.TokenValidator {
	return auth.TokenValidatorFunc(func(token string, claims jwt.Claims) bool {
		exists, err := storage.Exists(token)

		if err != nil {
			log.Printf("Error using token storage: %+v\n", err)
			return false
		}

		if exists {
			return false
		}

		if c, ok := claims.(*jwt.StandardClaims); ok {
			u := repo.UserByLogin(data.Login(c.Subject))
			err := u.Err()

			if err != nil {
				if err != content.ErrNoContent {
					log.Printf("Error getting user %s from repo: %+v\n", c.Subject, err)
				}

				return false
			}
		}

		return true
	})
}

func initAdminUser(repo content.Repo, secret []byte) error {
	users := repo.AllUsers()
	if repo.HasErr() {
		return errors.Wrap(repo.Err(), "getting all users")
	}

	if len(users) > 0 {
		return nil
	}

	u := repo.User()
	u.Data(data.User{Login: data.Login("admin"), Active: true, Admin: true})
	u.Password("admin", secret)
	u.Update()

	if u.HasErr() {
		return errors.Wrap(u.Err(), "updating user")
	}

	return nil
}

func readJSON(w http.ResponseWriter, r io.Reader, data interface{}) (stop bool) {
	if b, err := ioutil.ReadAll(r); err == nil {
		if err = json.Unmarshal(b, data); err != nil {
			http.Error(w, "Error decoding JSON request: "+err.Error(), http.StatusBadRequest)
			return true
		}
	} else {
		http.Error(w, "Error reading request body: "+err.Error(), http.StatusInternalServerError)
		return true
	}

	return false
}
