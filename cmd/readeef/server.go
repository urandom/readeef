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

	"github.com/alexedwards/scs/engine/boltstore"
	"github.com/alexedwards/scs/engine/memstore"
	"github.com/alexedwards/scs/session"
	"github.com/boltdb/bolt"
	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/api"
	"github.com/urandom/readeef/config"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/extract"
	"github.com/urandom/readeef/content/monitor"
	"github.com/urandom/readeef/content/processor"
	"github.com/urandom/readeef/content/repo"
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

func runServer(config config.Config, args []string) error {
	fs, err := readeef.NewFileSystem()
	if err != nil {
		return errors.WithMessage(err, "creating readeef filesystem")
	}

	/*
		sessDb, engine, err := initSessionEngine(config.Auth)
		if err != nil {
			return errors.WithMessage(err, "creating session engine")
		}
		defer sessDb.Close()
	*/
	engine := memstore.New(5 * time.Minute)

	log := initLog(config.Log)

	handler, err := web.Mux(fs, engine, config, log)
	if err != nil {
		return errors.WithMessage(err, "creating web mux")
	}

	mux := chi.NewRouter()
	mux.Mount("/", handler)

	service, err := sql.NewService(config.DB.Driver, config.DB.Connect, log)
	if err != nil {
		return errors.WithMessage(err, "creating content service")
	}

	if err = initAdminUser(service.UserRepo(), []byte(config.Auth.Secret)); err != nil {
		return errors.WithMessage(err, "initializing admin user")
	}

	feedManager := readeef.NewFeedManager(service.FeedRepo(), config, log)

	if processors, err := initFeedProcessors(config.FeedParser.Processors, config.FeedParser.ProxyHTTPURLTemplate, log); err == nil {
		for _, p := range processors {
			feedManager.AddFeedProcessor(p)
		}
	} else {
		return errors.WithMessage(err, "initializing parser processors")
	}

	searchProvider := initSearchProvider(config.Content, service, log)

	extractor, err := initArticleExtractor(config.Content, fs)
	if err != nil {
		return errors.WithMessage(err, "initializing content extract generator")
	}

	articleProcessors, err := initArticleProcessors(config.Content.Article.Processors, config.Content.Article.ProxyHTTPURLTemplate, log)
	if err != nil {
		return errors.WithMessage(err, "initializing article processors")
	}

	thumbnailer, err := initThumbnailGenerator(service, config.Content, extractor, articleProcessors, log)
	if err != nil {
		return errors.Wrap(err, "initializing thumbnail generator")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	initPopularityScore(ctx, service, config.Popularity, log)

	monitors := initFeedMonitors(ctx, config.FeedManager, service, searchProvider, thumbnailer, log)
	for _, m := range monitors {
		feedManager.AddFeedMonitor(m)
	}

	hubbub, err := initHubbub(config, service, monitors, feedManager, log)
	if err != nil {
		return errors.WithMessage(err, "initializing hubbub")
	}

	if hubbub != nil {
		feedManager.SetHubbub(hubbub)
	}

	handler, err = api.Mux(ctx, service, feedManager, hubbub, searchProvider, extractor, fs, articleProcessors, config, log)
	if err != nil {
		return errors.WithMessage(err, "creating api mux")
	}

	mux.Mount("/api", handler)

	feedManager.Start(ctx)

	server := makeHTTPServer(mux)

	if serverDevelPort > 0 {
		server.Addr = fmt.Sprintf(":%d", serverDevelPort)

		log.Infof("Starting server on address %s", server.Addr)
		if err = server.ListenAndServe(); err != nil {
			return errors.Wrap(err, "starting devel server")
		}
	}

	server.Addr = fmt.Sprintf("%s:%d", config.Server.Address, config.Server.Port)
	log.Infof("Starting server on address %s", server.Addr)

	if config.Server.AutoCert.Host != "" {
		if err := os.MkdirAll(config.Server.AutoCert.StorageDir, 0777); err != nil {
			return errors.Wrapf(err, "creating autocert storage dir %s", config.Server.AutoCert.StorageDir)
		}

		hostPolicy := func(ctx context.Context, host string) error {
			if host == config.Server.AutoCert.Host {
				return nil
			}
			return errors.Errorf("acme/autocert: only %s host is allowed", config.Server.AutoCert.Host)
		}

		m := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: hostPolicy,
			Cache:      autocert.DirCache(config.Server.AutoCert.StorageDir),
		}

		server.TLSConfig = &tls.Config{GetCertificate: m.GetCertificate}

		if err := server.ListenAndServeTLS("", ""); err != nil {
			return errors.Wrap(err, "starting auto-cert server")
		}
	} else if config.Server.CertFile != "" && config.Server.KeyFile != "" {
		if err := server.ListenAndServeTLS(config.Server.CertFile, config.Server.KeyFile); err != nil {
			return errors.Wrap(err, "starting tls server")
		}
	} else {
		if err = server.ListenAndServe(); err != nil {
			return errors.Wrap(err, "starting server")
		}
	}

	return nil
}

func initSessionEngine(config config.Auth) (*bolt.DB, session.Engine, error) {
	if err := os.MkdirAll(filepath.Dir(config.SessionStoragePath), 0777); err != nil {
		return nil, nil, errors.Wrapf(err, "creating session storage path %s", config.SessionStoragePath)
	}

	db, err := bolt.Open(config.SessionStoragePath, 0600, nil)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "opening bolt db with path: %s", config.SessionStoragePath)
	}

	return db, boltstore.New(db, 30), nil
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
	service repo.Service,
	searchProvider search.Provider,
	thumbnailer thumbnail.Generator,
	log log.Log,
) []monitor.Feed {
	monitors := []monitor.Feed{monitor.NewUnread(ctx, service, log)}

	for _, m := range config.Monitors {
		switch m {
		case "index":
			if searchProvider != nil {
				monitors = append(monitors, monitor.NewIndex(service.ArticleRepo(), searchProvider, log))
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
	service repo.Service,
	monitors []monitor.Feed,
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
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
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
