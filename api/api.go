package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	"github.com/urandom/handler/auth"
	"github.com/urandom/handler/method"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/api/fever"
	"github.com/urandom/readeef/api/ttrss"
	"github.com/urandom/readeef/config"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base/extractor"
	"github.com/urandom/readeef/content/base/monitor"
	contentProcessor "github.com/urandom/readeef/content/base/processor"
	"github.com/urandom/readeef/content/base/search"
	"github.com/urandom/readeef/content/base/thumbnailer"
	"github.com/urandom/readeef/content/base/token"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/parser"
	"github.com/urandom/readeef/parser/processor"
)

func Prepare(ctx context.Context, fs http.FileSystem, config config.Config, log readeef.Logger) (http.Handler, error) {
	repo, err := repo.New(config.DB.Driver, config.DB.Connect, log)
	if err != nil {
		return nil, errors.Wrap(err, "creating repo")
	}

	languageSupport := false
	if languages, err := readeef.GetLanguages(fs); err == nil {
		languageSupport = len(languages) > 0
	}

	features := features{
		I18N:       languageSupport,
		Popularity: len(config.Popularity.Providers) > 0,
	}

	if processors, err := initArticleProcessors(config.Content.ArticleProcessors, config.Content.ProxyHTTPURLTemplate, &features, log); err == nil {
		repo.ArticleProcessors(processors)
	} else {
		return nil, errors.Wrap(err, "initializing article processors")
	}

	if err := initAdminUser(repo, []byte(config.Auth.Secret)); err != nil {
		return nil, errors.Wrap(err, "initializing admin user")
	}

	feedManager := readeef.NewFeedManager(repo, config, log)

	if processors, err := initParserProcessors(config.FeedParser.Processors, config.FeedParser.ProxyHTTPURLTemplate, &features, log); err == nil {
		feedManager.ParserProcessors(processors)
	} else {
		return nil, errors.Wrap(err, "initializing parser processors")
	}

	searchProvider := initSearchProvider(config.Content, repo, log)

	if searchProvider != nil {
		features.Search = true
	}

	extractor, err := initContentExtractor(config.Content)
	if err != nil {
		return nil, errors.Wrap(err, "initializing content extractor")
	}

	if extractor != nil {
		features.Extractor = true
	}

	thumbnailer, err := initThumbnailer(config.Content, extractor, log)
	if err != nil {
		return nil, errors.Wrap(err, "initializing thumbnailer")
	}

	monitors := initFeedMonitors(config.FeedManager, repo, searchProvider, thumbnailer, &features, log)
	for _, m := range monitors {
		feedManager.AddFeedMonitor(m)
	}

	storage, err := initTokenStorage(config.Auth)
	if err != nil {
		return nil, errors.Wrap(err, "initializing token storage")
	}

	hubbub, err := initHubbub(config, repo, monitors, feedManager, log)
	if err != nil {
		return nil, errors.Wrap(err, "initializing hubbub")
	}

	routes := []routes{tokenRoutes(repo, storage, []byte(config.Auth.Secret), log)}

	if hubbub != nil {
		feedManager.Hubbub(hubbub)
		routes = append(routes, hubbubRoutes(hubbub, repo, feedManager, log))
	}

	emulatorRoutes := emulatorRoutes(ctx, repo, searchProvider, feedManager, config, log)
	routes = append(routes, emulatorRoutes...)

	routes = append(routes, mainRoutes(
		userMiddleware(repo, storage, []byte(config.Auth.Secret), log),
		featureRoutes(features),
		feedsRoutes(feedManager),
		articlesRoutes(extractor, searchProvider),
		opmlRoutes(feedManager),
		eventsRoutes(ctx, storage, feedManager),
		userRoutes([]byte(config.Auth.Secret)),
	))

	handler := Mux(routes)

	feedManager.Start()

	return handler, nil
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

func initSearchProvider(config config.Content, repo content.Repo, log readeef.Logger) content.SearchProvider {
	var searchProvider content.SearchProvider
	var err error

	switch config.SearchProvider {
	case "elastic":
		if searchProvider, err = search.NewElastic(
			config.ElasticURL,
			config.SearchBatchSize,
			log,
		); err != nil {
			log.Printf("Error initializing Elastic search: %v\n", err)
		}
	case "bleve":
		fallthrough
	default:
		if searchProvider, err = search.NewBleve(
			config.BlevePath,
			config.SearchBatchSize,
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

func initContentExtractor(config config.Content) (content.Extractor, error) {
	switch config.Extractor {
	case "readability":
		if ce, err := extractor.NewReadability(config.ReadabilityKey); err == nil {
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

func initThumbnailer(config config.Content, ce content.Extractor, log readeef.Logger) (content.Thumbnailer, error) {
	switch config.Thumbnailer {
	case "extract":
		if t, err := thumbnailer.NewExtract(ce, log); err == nil {
			return t, nil
		} else {
			return nil, errors.Wrap(err, "initializing Extract thumbnailer")
		}
	case "description":
		fallthrough
	default:
		return thumbnailer.NewDescription(log), nil
	}
}

func initFeedMonitors(
	config config.FeedManager,
	repo content.Repo,
	searchProvider content.SearchProvider,
	thumbnailer content.Thumbnailer,
	features *features,
	log readeef.Logger,
) []content.FeedMonitor {
	monitors := []content.FeedMonitor{monitor.NewUnread(repo, log)}

	for _, m := range config.Monitors {
		switch m {
		case "index":
			if searchProvider != nil {
				monitors = append(monitors, monitor.NewIndex(searchProvider, log))
				features.Search = true
			}
		case "thumbnailer":
			if thumbnailer != nil {
				monitors = append(monitors, monitor.NewThumbnailer(thumbnailer, log))
			}
		}
	}

	return monitors
}

func initHubbub(
	config config.Config,
	repo content.Repo,
	monitors []content.FeedMonitor,
	feedManager *readeef.FeedManager,
	log readeef.Logger,
) (*readeef.Hubbub, error) {
	if config.Hubbub.CallbackURL != "" {
		hubbub := readeef.NewHubbub(repo, config, log, "/api/v2/hubbub",
			feedManager.RemoveFeedChannel())

		if err := hubbub.InitSubscriptions(); err != nil {
			return nil, errors.Wrap(err, "initializing hubbub subscriptions")
		}

		hubbub.FeedMonitors(monitors)

		return hubbub, nil
	}

	return nil, nil
}

func initTokenStorage(config config.Auth) (content.TokenStorage, error) {
	return token.NewBoltStorage(config.TokenStoragePath)
}

type routes struct {
	path  string
	route func(r chi.Router)
}

func tokenRoutes(repo content.Repo, storage content.TokenStorage, secret []byte, log readeef.Logger) routes {
	return routes{path: "/v2/token", route: func(r chi.Router) {
		r.Method(method.POST, "/", tokenCreate(repo, secret, log))
		r.Method(method.DELETE, "/", tokenDelete(storage, secret, log))
	}}
}

func hubbubRoutes(hubbub *readeef.Hubbub, repo content.Repo, feedManager *readeef.FeedManager, log readeef.Logger) routes {
	handler := hubbubRegistration(hubbub, repo, feedManager.AddFeedChannel(), feedManager.RemoveFeedChannel(), log)

	return routes{path: "/v2/hubbub", route: func(r chi.Router) {
		r.Get("/", handler)
		r.Post("/", handler)
	}}
}

func emulatorRoutes(
	ctx context.Context,
	repo content.Repo,
	searchProvider content.SearchProvider,
	feedManager *readeef.FeedManager,
	config config.Config,
	log readeef.Logger,
) []routes {
	rr := make([]routes, 0, len(config.API.Emulators))

	for _, e := range config.API.Emulators {
		switch e {
		case "tt-rss":
			rr = append(rr, routes{
				path: fmt.Sprintf("/v%d/tt-rss/", ttrss.API_LEVEL),
				route: func(r chi.Router) {
					r.Get("/", ttrss.FakeWebHandler)

					r.Post("/api/", ttrss.Handler(
						ctx, repo, searchProvider, feedManager,
						[]byte(config.Auth.Secret), config.FeedManager.Converted.UpdateInterval,
						log,
					))
				},
			})
		case "fever":
			rr = append(rr, routes{
				path: fmt.Sprintf("/v%d/fever/", fever.API_VERSION),
				route: func(r chi.Router) {
					r.Post("/", fever.Handler(repo, log))
				},
			})
		}
	}

	return rr
}

type middleware func(next http.Handler) http.Handler

func mainRoutes(middleware []middleware, subroutes ...routes) routes {
	return routes{path: "/v2", route: func(r chi.Router) {
		for _, m := range middleware {
			r.Use(m)
		}

		for _, sub := range subroutes {
			r.Route(sub.path, sub.route)
		}
	}}
}

func userMiddleware(repo content.Repo, storage content.TokenStorage, secret []byte, log readeef.Logger) []middleware {
	return []middleware{
		func(next http.Handler) http.Handler {
			return auth.RequireToken(next, tokenValidator(repo, storage, log), secret)
		},
		func(next http.Handler) http.Handler {
			return userContext(repo, next, log)
		},
		userValidator,
	}
}

func featureRoutes(features features) routes {
	return routes{path: "/features", route: func(r chi.Router) {
		r.Get("/", featuresHandler(features))
	}}
}

func feedsRoutes(feedManager *readeef.FeedManager) routes {
	return routes{path: "/feed", route: func(r chi.Router) {
		r.Get("/", listFeeds)
		r.Post("/", addFeed(feedManager))

		r.Get("/discover", discoverFeeds(feedManager))

		r.Route("/{feedId:[0-9]+}", func(r chi.Router) {
			r.Use(feedContext)

			r.Delete("/", deleteFeed(feedManager))

			r.Get("/tags", getFeedTags)
			r.Post("/tags", setFeedTags)

		})
	}}
}

func articlesRoutes(extractor content.Extractor, searchProvider content.SearchProvider) routes {
	return routes{path: "/article", route: func(r chi.Router) {
		r.Get("/", getArticles(userRepoType))

		if searchProvider != nil {
			r.Route("/search", func(r chi.Router) {
				r.Get("/*", articleSearch(searchProvider, userRepoType))
				r.With(feedContext).Get("/feed/{feedId:[0-9]+}/*", articleSearch(searchProvider, feedRepoType))
				r.With(tagContext).Get("/tag/{tagId:[0-9]+}/*", articleSearch(searchProvider, tagRepoType))
			})
		}

		r.Post("/read", articlesReadStateChange(userRepoType))

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

	}}
}

func opmlRoutes(feedManager *readeef.FeedManager) routes {
	return routes{path: "/opml", route: func(r chi.Router) {
		r.Get("/", exportOPML(feedManager))
		r.Post("/", importOPML(feedManager))
	}}
}

func eventsRoutes(ctx context.Context, storage content.TokenStorage, feedManager *readeef.FeedManager) routes {
	return routes{path: "/events", route: func(r chi.Router) {
		r.Get("/", eventSocket(ctx, storage, feedManager))
	}}
}

func userRoutes(secret []byte) routes {
	return routes{path: "/user", route: func(r chi.Router) {
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
	}}
}

func Mux(routes []routes) http.Handler {
	r := chi.NewRouter()

	r.Route("/api", func(r chi.Router) {
		for _, sub := range routes {
			r.Route(sub.path, sub.route)
		}
	})

	return r
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
