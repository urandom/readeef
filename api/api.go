package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	"github.com/urandom/handler/auth"
	"github.com/urandom/handler/encoding"
	"github.com/urandom/handler/method"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/api/fever"
	"github.com/urandom/readeef/api/token"
	"github.com/urandom/readeef/api/ttrss"
	"github.com/urandom/readeef/config"
	"github.com/urandom/readeef/content/extract"
	"github.com/urandom/readeef/content/processor"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/content/repo/eventable"
	"github.com/urandom/readeef/content/search"
	"github.com/urandom/readeef/log"
)

type mw func(http.Handler) http.Handler

func Mux(
	ctx context.Context,
	service eventable.Service,
	feedManager *readeef.FeedManager,
	searchProvider search.Provider,
	extractor extract.Generator,
	fs http.FileSystem,
	processors []processor.Article,
	config config.Config,
	log log.Log,
	access mw,
) (http.Handler, error) {

	features := features{
		Popularity: len(config.Popularity.Providers) > 0,
		ProxyHTTP:  hasProxy(config),
		Search:     searchProvider != nil,
		Extractor:  extractor != nil,
	}

	var gzip mw = func(n http.Handler) http.Handler {
		return encoding.Gzip(n, encoding.Logger(log))
	}

	storage, err := initTokenStorage(config.Auth)
	if err != nil {
		return nil, errors.Wrap(err, "initializing token storage")
	}

	routes := []routes{tokenRoutes(service.UserRepo(), storage, []byte(config.Auth.Secret), log, gzip, access)}

	if config.Hubbub.CallbackURL != "" {
		routes = append(routes, hubbubRoutes(service, log, gzip, access))
	}

	emulatorRoutes := emulatorRoutes(ctx, service, searchProvider, feedManager, processors, config, log, gzip, access)
	routes = append(routes, emulatorRoutes...)

	routes = append(routes, mainRoutes(
		userMiddleware(service.UserRepo(), storage, []byte(config.Auth.Secret), log),
		featureRoutes(features, gzip, access),
		feedsRoutes(service, feedManager, log, gzip, access),
		tagRoutes(service.TagRepo(), log, gzip, access),
		articlesRoutes(service, extractor, searchProvider, processors, config, log, gzip, access),
		opmlRoutes(service, feedManager, log, gzip, access),
		eventsRoutes(ctx, service, storage, feedManager, log),
		userRoutes(service, []byte(config.Auth.Secret), log, gzip, access),
	))

	r := chi.NewRouter()

	r.Route("/v2", func(r chi.Router) {
		for _, sub := range routes {
			r.Route(sub.path, sub.route)
		}
	})

	return r, nil
}

func hasProxy(config config.Config) bool {
	for _, p := range config.Content.Article.Processors {
		if p == "proxy-http" {
			return true
		}
	}

	for _, p := range config.FeedParser.Processors {
		if p == "proxy-http" {
			return true
		}
	}

	return false
}

func initTokenStorage(config config.Auth) (token.Storage, error) {
	if err := os.MkdirAll(filepath.Dir(config.TokenStoragePath), 0777); err != nil {
		return nil, errors.Wrapf(err, "creating token storage path %s", config.TokenStoragePath)
	}

	return token.NewBoltStorage(config.TokenStoragePath)
}

type routes struct {
	path  string
	route func(r chi.Router)
}

func timeout(d time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, d, "")
	}
}

func tokenRoutes(repo repo.User, storage token.Storage, secret []byte, log log.Log, gzip, access mw) routes {
	return routes{path: "/token", route: func(r chi.Router) {
		r.Use(timeout(time.Second), gzip, access)
		r.Method(method.POST, "/", tokenCreate(repo, secret, log))
		r.Method(method.DELETE, "/", tokenDelete(storage, secret, log))
	}}
}

func hubbubRoutes(service repo.Service, log log.Log, gzip, access mw) routes {
	handler := hubbubRegistration(service, log)

	return routes{path: "/hubbub", route: func(r chi.Router) {
		r.Use(timeout(5*time.Second), gzip, access)
		r.Get("/", handler)
		r.Post("/", handler)
	}}
}

func emulatorRoutes(
	ctx context.Context,
	service repo.Service,
	searchProvider search.Provider,
	feedManager *readeef.FeedManager,
	processors []processor.Article,
	config config.Config,
	log log.Log,
	gzip, access mw,
) []routes {
	rr := make([]routes, 0, len(config.API.Emulators))

	for _, e := range config.API.Emulators {
		switch e {
		case "tt-rss":
			rr = append(rr, routes{
				path: "/tt-rss/",
				route: func(r chi.Router) {
					r.Use(timeout(10*time.Second), gzip, access)
					r.Get("/", ttrss.FakeWebHandler)

					r.Post("/api/", ttrss.Handler(
						ctx, service, searchProvider, feedManager, processors,
						[]byte(config.Auth.Secret), config.FeedManager.Converted.UpdateInterval,
						log,
					))
				},
			})
		case "fever":
			rr = append(rr, routes{
				path: "/fever/",
				route: func(r chi.Router) {
					r.Use(timeout(10*time.Second), gzip, access)
					r.Post("/", fever.Handler(service, processors, log))
				},
			})
		}
	}

	return rr
}

type middleware func(next http.Handler) http.Handler

func mainRoutes(middleware []middleware, subroutes ...routes) routes {
	return routes{path: "/", route: func(r chi.Router) {
		for _, m := range middleware {
			r.Use(m)
		}

		for _, sub := range subroutes {
			r.Route(sub.path, sub.route)
		}
	}}
}

func userMiddleware(repo repo.User, storage token.Storage, secret []byte, log log.Log) []middleware {
	return []middleware{
		func(next http.Handler) http.Handler {
			return auth.RequireToken(next, tokenValidator(repo, storage, log), secret)
		},
		func(next http.Handler) http.Handler {
			return userContext(repo, next, log)
		},
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var err error
				if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
					err = r.ParseMultipartForm(0)
				} else {
					err = r.ParseForm()
				}
				if err != nil {
					http.Error(w, "Error parsing form data", http.StatusBadRequest)
					return
				}
				next.ServeHTTP(w, r)
			})
		},
		userValidator,
	}
}

func featureRoutes(features features, gzip, access mw) routes {
	return routes{path: "/features", route: func(r chi.Router) {
		r.Use(timeout(time.Second), gzip, access)
		r.Get("/", featuresHandler(features))
	}}
}

func feedsRoutes(service repo.Service, feedManager *readeef.FeedManager, log log.Log, gzip, access mw) routes {
	return routes{path: "/feed", route: func(r chi.Router) {
		feedRepo := service.FeedRepo()
		r.Use(gzip, access)
		r.With(timeout(5*time.Second)).Get("/", listFeeds(feedRepo, log))
		r.With(timeout(15*time.Second)).Post("/", addFeed(feedRepo, feedManager))

		r.With(timeout(30*time.Second)).Get("/discover", discoverFeeds(feedRepo, feedManager, log))

		r.Route("/{feedID:[0-9]+}", func(r chi.Router) {
			r.Use(feedContext(service.FeedRepo(), log))
			r.Use(timeout(5 * time.Second))

			r.Delete("/", deleteFeed(feedRepo, feedManager, log))

			r.Get("/tags", getFeedTags(service.TagRepo(), log))
			r.Put("/tags", setFeedTags(feedRepo, log))

		})
	}}
}

func tagRoutes(repo repo.Tag, log log.Log, gzip, access mw) routes {
	return routes{path: "/tag", route: func(r chi.Router) {
		r.Use(timeout(5*time.Second), gzip, access)
		r.Get("/", listTags(repo, log))
		r.Get("/feedIDs", getTagsFeedIDs(repo, log))

		r.Route("/{tagID:[0-9]+}", func(r chi.Router) {
			r.Use(tagContext(repo, log))

			r.Get("/feedIDs", getTagFeedIDs(repo, log))
		})
	}}
}

func articlesRoutes(
	service repo.Service,
	extractor extract.Generator,
	searchProvider search.Provider,
	processors []processor.Article,
	config config.Config,
	log log.Log,
	gzip, access mw,
) routes {
	articleRepo := service.ArticleRepo()
	feedRepo := service.FeedRepo()
	tagRepo := service.TagRepo()

	return routes{path: "/article", route: func(r chi.Router) {
		r.Use(timeout(30*time.Second), gzip, access)
		r.Get("/", getArticles(service, userRepoType, noRepoType, processors, config.API.Limits.ArticlesPerQuery, log))

		if searchProvider != nil {
			r.Route("/search", func(r chi.Router) {
				r.Get("/",
					articleSearch(service, searchProvider, userRepoType, processors, config.API.Limits.ArticlesPerQuery, log))
				r.With(feedContext(feedRepo, log)).Get("/feed/{feedID:[0-9]+}",
					articleSearch(service, searchProvider, feedRepoType, processors, config.API.Limits.ArticlesPerQuery, log))
				r.With(tagContext(tagRepo, log)).Get("/tag/{tagID:[0-9]+}",
					articleSearch(service, searchProvider, tagRepoType, processors, config.API.Limits.ArticlesPerQuery, log))
			})
		}

		r.Get("/ids", getIDs(service, userRepoType, noRepoType, config.API.Limits.ArticlesPerQuery, log))

		r.Post("/read", articlesStateChange(service, userRepoType, read, log))
		r.Delete("/read", articlesStateChange(service, userRepoType, read, log))

		r.Route("/{articleID:[0-9]+}", func(r chi.Router) {
			r.Use(articleContext(articleRepo, processors, log))

			r.Get("/", getArticle)
			if extractor != nil {
				r.Get("/format", formatArticle(service.ExtractRepo(), extractor, processors, log))
			}
			r.Post("/read", articleStateChange(articleRepo, read, log))
			r.Delete("/read", articleStateChange(articleRepo, read, log))
			r.Post("/favorite", articleStateChange(articleRepo, favorite, log))
			r.Delete("/favorite", articleStateChange(articleRepo, favorite, log))
		})

		r.Route("/favorite", func(r chi.Router) {
			r.Get("/", getArticles(service, favoriteRepoType, noRepoType, processors, config.API.Limits.ArticlesPerQuery, log))
			r.Post("/", articlesStateChange(service, userRepoType, favorite, log))
			r.Delete("/", articlesStateChange(service, userRepoType, favorite, log))

			r.Get("/ids", getIDs(service, favoriteRepoType, noRepoType, config.API.Limits.ArticlesPerQuery, log))

			r.Post("/read", articlesStateChange(service, favoriteRepoType, read, log))
			r.Delete("/read", articlesStateChange(service, favoriteRepoType, read, log))
		})

		r.Route("/popular", func(r chi.Router) {

			r.Route("/feed/{feedID:[0-9]+}", func(r chi.Router) {
				r.Use(feedContext(feedRepo, log))

				r.Get("/", getArticles(service, popularRepoType, feedRepoType, processors, config.API.Limits.ArticlesPerQuery, log))
				r.Get("/ids", getIDs(service, popularRepoType, feedRepoType, config.API.Limits.ArticlesPerQuery, log))
			})

			r.Route("/tag/{tagID:[0-9]+}", func(r chi.Router) {
				r.Use(tagContext(tagRepo, log))

				r.Get("/", getArticles(service, popularRepoType, tagRepoType, processors, config.API.Limits.ArticlesPerQuery, log))
				r.Get("/ids", getIDs(service, popularRepoType, tagRepoType, config.API.Limits.ArticlesPerQuery, log))
			})

			r.Get("/", getArticles(service, popularRepoType, userRepoType, processors, config.API.Limits.ArticlesPerQuery, log))
			r.Get("/ids", getIDs(service, popularRepoType, userRepoType, config.API.Limits.ArticlesPerQuery, log))
		})

		r.Route("/feed/{feedID:[0-9]+}", func(r chi.Router) {
			r.Use(feedContext(feedRepo, log))

			r.Get("/", getArticles(service, feedRepoType, noRepoType, processors, config.API.Limits.ArticlesPerQuery, log))
			r.Get("/ids", getIDs(service, feedRepoType, noRepoType, config.API.Limits.ArticlesPerQuery, log))

			r.Post("/read", articlesStateChange(service, feedRepoType, read, log))
			r.Delete("/read", articlesStateChange(service, feedRepoType, read, log))
		})

		r.Route("/tag/{tagID:[0-9]+}", func(r chi.Router) {
			r.Use(tagContext(tagRepo, log))

			r.Get("/", getArticles(service, tagRepoType, noRepoType, processors, config.API.Limits.ArticlesPerQuery, log))
			r.Get("/ids", getIDs(service, tagRepoType, noRepoType, config.API.Limits.ArticlesPerQuery, log))

			r.Post("/read", articlesStateChange(service, tagRepoType, read, log))
			r.Delete("/read", articlesStateChange(service, tagRepoType, read, log))
		})

	}}
}

func opmlRoutes(service repo.Service, feedManager *readeef.FeedManager, log log.Log, gzip, access mw) routes {
	return routes{path: "/opml", route: func(r chi.Router) {
		r.Use(gzip, access)
		r.With(timeout(10*time.Second)).Get("/", exportOPML(service, log))
		r.With(timeout(30*time.Second)).Post("/", importOPML(service.FeedRepo(), feedManager, log))
	}}
}

func eventsRoutes(
	ctx context.Context,
	service eventable.Service,
	storage token.Storage,
	feedManager *readeef.FeedManager,
	log log.Log,
) routes {
	return routes{path: "/events", route: func(r chi.Router) {
		r.Get("/", eventSocket(ctx, service, storage, log))
	}}
}

func userRoutes(service repo.Service, secret []byte, log log.Log, gzip, access mw) routes {
	repo := service.UserRepo()
	return routes{path: "/user", route: func(r chi.Router) {
		r.Use(timeout(5*time.Second), gzip, access)

		r.Get("/current", getUserData)
		r.Post("/token", createUserToken(secret, log))

		r.Route("/settings", func(r chi.Router) {
			r.Get("/", getSettingKeys)
			r.Get("/{key}", getSettingValue)
			r.Put("/{key}", setSettingValue(repo, secret, log))
		})

		r.Route("/", func(r chi.Router) {
			r.Use(adminValidator)

			r.Get("/", listUsers(repo, log))

			r.Post("/", addUser(repo, secret, log))
			r.Route("/{name}", func(r chi.Router) {
				r.Delete("/", deleteUser(repo, log))
			})

			r.Put("/{name}/settings/{key}", setSettingValue(repo, secret, log))
		})
	}}
}

func fatal(w http.ResponseWriter, log log.Log, format string, err error) {
	log.Printf(format, err)
	http.Error(w, fmt.Sprintf(format, err.Error()), http.StatusInternalServerError)
}
