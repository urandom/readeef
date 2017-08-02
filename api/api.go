package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	"github.com/urandom/handler/auth"
	"github.com/urandom/handler/method"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/api/fever"
	"github.com/urandom/readeef/api/ttrss"
	"github.com/urandom/readeef/config"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base/token"
	"github.com/urandom/readeef/content/extract"
	"github.com/urandom/readeef/content/search"
)

func Mux(
	ctx context.Context,
	repo content.Repo,
	feedManager *readeef.FeedManager,
	hubbub *readeef.Hubbub,
	searchProvider search.Provider,
	extractor extract.Generator,
	fs http.FileSystem,
	config config.Config,
	log readeef.Logger,
) (http.Handler, error) {

	languageSupport := false
	if languages, err := readeef.GetLanguages(fs); err == nil {
		languageSupport = len(languages) > 0
	}

	features := features{
		I18N:       languageSupport,
		Popularity: len(config.Popularity.Providers) > 0,
		ProxyHTTP:  hasProxy(config),
		Search:     searchProvider != nil,
		Extractor:  extractor != nil,
	}

	storage, err := initTokenStorage(config.Auth)
	if err != nil {
		return nil, errors.Wrap(err, "initializing token storage")
	}

	routes := []routes{tokenRoutes(repo, storage, []byte(config.Auth.Secret), log)}

	if hubbub != nil {
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

	r := chi.NewRouter()

	r.Route("/v2", func(r chi.Router) {
		for _, sub := range routes {
			r.Route(sub.path, sub.route)
		}
	})

	return r, nil
}

func hasProxy(config config.Config) bool {
	for _, p := range config.Content.ArticleProcessors {
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

func initTokenStorage(config config.Auth) (content.TokenStorage, error) {
	if err := os.MkdirAll(filepath.Dir(config.TokenStoragePath), 0777); err != nil {
		return nil, errors.Wrapf(err, "creating token storage path %s", config.TokenStoragePath)
	}

	return token.NewBoltStorage(config.TokenStoragePath)
}

type routes struct {
	path  string
	route func(r chi.Router)
}

func tokenRoutes(repo content.Repo, storage content.TokenStorage, secret []byte, log readeef.Logger) routes {
	return routes{path: "/token", route: func(r chi.Router) {
		r.Method(method.POST, "/", tokenCreate(repo, secret, log))
		r.Method(method.DELETE, "/", tokenDelete(storage, secret, log))
	}}
}

func hubbubRoutes(hubbub *readeef.Hubbub, repo content.Repo, feedManager *readeef.FeedManager, log readeef.Logger) routes {
	handler := hubbubRegistration(hubbub, repo, feedManager, log)

	return routes{path: "/hubbub", route: func(r chi.Router) {
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
	return routes{path: "/", route: func(r chi.Router) {
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
