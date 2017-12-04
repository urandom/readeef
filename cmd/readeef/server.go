package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/acme/autocert"

	"github.com/alexedwards/scs"
	"github.com/alexedwards/scs/stores/boltstore"
	"github.com/alexedwards/scs/stores/memstore"
	"github.com/boltdb/bolt"
	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	"github.com/urandom/handler/encoding"
	handlerLog "github.com/urandom/handler/log"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/api"
	"github.com/urandom/readeef/config"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/extract"
	"github.com/urandom/readeef/content/monitor"
	"github.com/urandom/readeef/content/processor"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/content/repo/eventable"
	"github.com/urandom/readeef/content/repo/logging"
	"github.com/urandom/readeef/content/repo/sql"
	"github.com/urandom/readeef/content/search"
	"github.com/urandom/readeef/content/thumbnail"
	"github.com/urandom/readeef/log"
	"github.com/urandom/readeef/popularity"
	"github.com/urandom/readeef/web"
)

var (
	serverDevelPort int
)

func runServer(cfg config.Config, args []string) error {
	fs, err := readeef.NewFileSystem()
	if err != nil {
		return errors.WithMessage(err, "creating readeef filesystem")
	}

	if serverDevelPort > 0 {
		fs = http.Dir(".")
	}

	/*
		sessDb, engine, err := initSessionEngine(cfg.Auth)
		if err != nil {
			return errors.WithMessage(err, "creating session engine")
		}
		defer sessDb.Close()
	*/
	engine := memstore.New(5 * time.Minute)
	sessionManager := scs.NewManager(engine)
	sessionManager.Lifetime(240 * time.Hour)

	logger := initLog(cfg.Log)

	handler, err := web.Mux(fs, sessionManager, cfg, logger)
	if err != nil {
		return errors.WithMessage(err, "creating web mux")
	}

	var accessLog log.Log
	if cfg.Log.AccessFile != "" {
		if cfg.Log.AccessFile == cfg.Log.File {
			accessLog = logger
		} else {
			cfg := config.Log{File: cfg.Log.AccessFile}
			cfg.Convert()
			accessLog = log.WithStd(cfg)
		}
	}

	accessMiddleware := func(next http.Handler) http.Handler {
		if accessLog == nil {
			return next
		}
		return handlerLog.Access(next, handlerLog.Logger(accessLog))
	}

	mux := chi.NewRouter()
	mux.Mount("/", accessMiddleware(encoding.Gzip(handler)))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var baseService repo.Service
	baseService, err = sql.NewService(cfg.DB.Driver, cfg.DB.Connect, logger)
	if err != nil {
		return errors.WithMessage(err, "creating content service")
	}

	if cfg.Log.RepoCallDuration {
		baseService = logging.NewService(baseService, logger)
	}
	service := eventable.NewService(ctx, baseService, logger)

	if err = initAdminUser(service.UserRepo(), []byte(cfg.Auth.Secret)); err != nil {
		return errors.WithMessage(err, "initializing admin user")
	}

	feedManager := readeef.NewFeedManager(service.FeedRepo(), cfg, logger)

	if processors, err := initFeedProcessors(cfg.FeedParser.Processors, cfg.FeedParser.ProxyHTTPURLTemplate, logger); err == nil {
		for _, p := range processors {
			feedManager.AddFeedProcessor(p)
		}
	} else {
		return errors.WithMessage(err, "initializing parser processors")
	}

	searchProvider := initSearchProvider(cfg.Content, service, logger)

	extractor, err := initArticleExtractor(cfg.Content, fs)
	if err != nil {
		return errors.WithMessage(err, "initializing content extract generator")
	}

	articleProcessors, err := initArticleProcessors(cfg.Content.Article.Processors, cfg.Content.Article.ProxyHTTPURLTemplate, logger)
	if err != nil {
		return errors.WithMessage(err, "initializing article processors")
	}

	thumbnailer, err := initThumbnailGenerator(service, cfg.Content, extractor, articleProcessors, logger)
	if err != nil {
		return errors.Wrap(err, "initializing thumbnail generator")
	}

	initPopularityScore(ctx, service, cfg.Popularity, logger)

	initFeedMonitors(ctx, cfg.FeedManager, service, searchProvider, thumbnailer, logger)

	hubbub, err := initHubbub(cfg, service, feedManager, logger)
	if err != nil {
		return errors.WithMessage(err, "initializing hubbub")
	}

	if hubbub != nil {
		feedManager.SetHubbub(hubbub)
	}

	handler, err = api.Mux(ctx, service, feedManager, searchProvider, extractor, fs, articleProcessors, cfg, logger, accessMiddleware)
	if err != nil {
		return errors.WithMessage(err, "creating api mux")
	}

	if serverDevelPort > 0 {
		handler = cors.New(cors.Options{
			ExposedHeaders: []string{"Authorization"},
			AllowedHeaders: []string{"Authorization"},
			AllowedOrigins: []string{"*"},
		}).Handler(handler)
	}

	mux.Mount("/api", handler)

	feedManager.Start(ctx)

	server := makeHTTPServer(mux)

	if serverDevelPort > 0 {
		server.Addr = fmt.Sprintf(":%d", serverDevelPort)

		logger.Infof("Starting server on address %s", server.Addr)
		if err = server.ListenAndServe(); err != nil {
			return errors.Wrap(err, "starting devel server")
		}

		return nil
	}

	server.Addr = fmt.Sprintf("%s:%d", cfg.Server.Address, cfg.Server.Port)
	logger.Infof("Starting server on address %s", server.Addr)

	if cfg.Server.AutoCert.Host != "" {
		if err := os.MkdirAll(cfg.Server.AutoCert.StoragePath, 0777); err != nil {
			return errors.Wrapf(err, "creating autocert storage dir %s", cfg.Server.AutoCert.StoragePath)
		}

		hostPolicy := func(ctx context.Context, host string) error {
			if host == cfg.Server.AutoCert.Host {
				return nil
			}
			return errors.Errorf("acme/autocert: only %s host is allowed", cfg.Server.AutoCert.Host)
		}

		m := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: hostPolicy,
			Cache:      autocert.DirCache(cfg.Server.AutoCert.StoragePath),
		}

		server.TLSConfig = &tls.Config{GetCertificate: m.GetCertificate}

		if err := server.ListenAndServeTLS("", ""); err != nil {
			return errors.Wrap(err, "starting auto-cert server")
		}
	} else if cfg.Server.CertFile != "" && cfg.Server.KeyFile != "" {
		if err := server.ListenAndServeTLS(cfg.Server.CertFile, cfg.Server.KeyFile); err != nil {
			return errors.Wrap(err, "starting tls server")
		}
	} else {
		if err = server.ListenAndServe(); err != nil {
			return errors.Wrap(err, "starting server")
		}
	}

	return nil
}

func initSessionEngine(config config.Auth) (*bolt.DB, scs.Store, error) {
	if err := os.MkdirAll(filepath.Dir(config.SessionStoragePath), 0777); err != nil {
		return nil, nil, errors.Wrapf(err, "creating session storage path %s", config.SessionStoragePath)
	}

	db, err := bolt.Open(config.SessionStoragePath, 0600, nil)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "opening bolt db with path: %s", config.SessionStoragePath)
	}

	store := boltstore.New(db, 30)
	return db, store, nil
}

func initLog(config config.Log) log.Log {
	return log.WithLogrus(config)
}

func initAdminUser(repo repo.User, secret []byte) error {
	users, err := repo.All()
	if err != nil {
		return errors.WithMessage(err, "getting all users")
	}

	if len(users) > 0 {
		return nil
	}

	u := content.User{Login: "admin", Active: true, Admin: true}
	u.Password("admin", secret)

	if err = repo.Update(u); err != nil {
		return errors.WithMessage(err, "updating user")
	}

	return nil
}

func initArticleProcessors(names []string, proxyTemplate string, log log.Log) ([]processor.Article, error) {
	var processors []processor.Article

	for _, p := range names {
		switch p {
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
			}
		case "insert-thumbnail-target":
			processors = append(processors, processor.NewInsertThumbnailTarget(log))
		case "unescape":
			processors = append(processors, processor.NewUnescape(log))
		}
	}

	return processors, nil
}

func initFeedProcessors(names []string, proxyTemplate string, log log.Log) ([]processor.Feed, error) {
	var processors []processor.Feed

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
			}
		case "cleanup":
			processors = append(processors, processor.NewCleanup(log))
		case "top-image-marker":
			processors = append(processors, processor.NewTopImageMarker(log))
		case "unescape":
			processors = append(processors, processor.NewUnescape(log))
		}
	}

	return processors, nil
}

func initSearchProvider(config config.Content, service repo.Service, log log.Log) search.Provider {
	var searchProvider search.Provider
	var err error

	switch config.Search.Provider {
	case "elastic":
		if searchProvider, err = search.NewElastic(
			config.Search.ElasticURL,
			config.Search.BatchSize,
			service,
			log,
		); err != nil {
			log.Printf("Error initializing Elastic search: %+v\n", err)
			searchProvider = nil
		}
	case "bleve":
		fallthrough
	default:
		if searchProvider, err = search.NewBleve(
			config.Search.BlevePath,
			config.Search.BatchSize,
			service,
			log,
		); err != nil {
			log.Printf("Error initializing Bleve search: %+v\n", err)
			searchProvider = nil
		}
	}

	if searchProvider != nil {
		if searchProvider.IsNewIndex() {
			go func() {
				if err := search.Reindex(searchProvider, service.ArticleRepo()); err != nil {
					log.Printf("Error reindexing all articles: %+v", err)
				}
			}()
		}
	}

	return searchProvider
}

func initArticleExtractor(config config.Content, fs http.FileSystem) (extract.Generator, error) {
	switch config.Extract.Generator {
	case "readability":
		if ce, err := extract.WithReadability(config.Extract.ReadabilityKey); err == nil {
			return ce, nil
		} else {
			return nil, errors.WithMessage(err, "initializing Readability extract generator")
		}
	case "goose":
		fallthrough
	default:
		if ce, err := extract.WithGoose("templates", fs); err == nil {
			return ce, nil
		} else {
			return nil, errors.WithMessage(err, "initializing Goose extract generator")
		}
	}
}

func initThumbnailGenerator(
	service repo.Service,
	config config.Content,
	extract extract.Generator,
	processors []processor.Article,
	log log.Log,
) (thumbnail.Generator, error) {

	switch config.ThumbnailGenerator {
	case "extract":
		if t, err := thumbnail.FromExtract(service.ThumbnailRepo(), service.ExtractRepo(), extract, processors, log); err == nil {
			return t, nil
		} else {
			return nil, errors.WithMessage(err, "initializing Extract thumbnail generator")
		}
	case "description":
		fallthrough
	default:
		return thumbnail.FromDescription(service.ThumbnailRepo(), log), nil
	}
}

func initPopularityScore(ctx context.Context, service repo.Service, config config.Popularity, log log.Log) {
	popularity.New(config, log).ScoreContent(ctx, service)
}

func initFeedMonitors(
	ctx context.Context,
	config config.FeedManager,
	service eventable.Service,
	searchProvider search.Provider,
	thumbnailer thumbnail.Generator,
	log log.Log,
) {
	go monitor.Unread(ctx, service, log)
	go monitor.UserFilters(service, log)

	for _, m := range config.Monitors {
		switch m {
		case "index":
			if searchProvider != nil {
				go monitor.Index(service, searchProvider, log)
			}
		case "thumbnailer":
			if thumbnailer != nil {
				go monitor.Thumbnailer(service, thumbnailer, log)
			}
		}
	}
}

func initHubbub(
	config config.Config,
	service repo.Service,
	feedManager *readeef.FeedManager,
	log log.Log,
) (*readeef.Hubbub, error) {
	if config.Hubbub.CallbackURL != "" {
		hubbub := readeef.NewHubbub(service, config, log, "/api/v2/hubbub", feedManager)

		if err := hubbub.InitSubscriptions(); err != nil {
			return nil, errors.WithMessage(err, "initializing hubbub subscriptions")
		}

		return hubbub, nil
	}

	return nil, nil
}

func makeHTTPServer(mux http.Handler) *http.Server {
	return &http.Server{
		ReadTimeout: 5 * time.Second,
		// Disable the global Write Timeout. Timeout on a
		// per-request to preserve SSE connections.
		WriteTimeout: 0,
		IdleTimeout:  120 * time.Second,
		Handler:      mux,
	}
}

func init() {
	flags := flag.NewFlagSet("server", flag.ExitOnError)
	flags.IntVar(&serverDevelPort, "devel-port", 0, "when specified, runs an http server on that port")

	commands = append(commands, Command{
		Name:  "server",
		Desc:  "feed aggregator server",
		Flags: flags,
		Run:   runServer,
	})
}
