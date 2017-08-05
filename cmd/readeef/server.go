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
	"github.com/alexedwards/scs/session"
	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/api"
	"github.com/urandom/readeef/config"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/extract"
	"github.com/urandom/readeef/content/monitor"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/content/repo/sql"
	"github.com/urandom/readeef/content/search"
	"github.com/urandom/readeef/content/thumbnail"
	"github.com/urandom/readeef/parser"
	"github.com/urandom/readeef/parser/processor"
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

	sessDb, engine, err := initSessionEngine(config.Auth)
	if err != nil {
		return errors.WithMessage(err, "creating session engine")
	}
	defer sessDb.Close()

	handler, err := web.Mux(fs, engine, config)
	if err != nil {
		return errors.WithMessage(err, "creating web mux")
	}

	mux := http.NewServeMux()
	mux.Handle("/", handler)

	log := initLog(config.Log)

	service, err := sql.NewService(config.DB.Driver, config.DB.Connect, log)
	if err != nil {
		return errors.WithMessage(err, "creating content service")
	}

	if err = initAdminUser(service.UserRepo(), []byte(config.Auth.Secret)); err != nil {
		return errors.WithMessage(err, "initializing admin user")
	}

	feedManager := readeef.NewFeedManager(service.FeedRepo(), config, log)

	if processors, err := initParserProcessors(config.FeedParser.Processors, config.FeedParser.ProxyHTTPURLTemplate, log); err == nil {
		for _, p := range processors {
			feedManager.AddParserProcessor(p)
		}
	} else {
		return errors.WithMessage(err, "initializing parser processors")
	}

	searchProvider := initSearchProvider(config.Content, service.ArticleRepo(), log)

	extractor, err := initArticleExtractor(config.Content, fs)
	if err != nil {
		return errors.WithMessage(err, "initializing content extract generator")
	}

	thumbnailer, err := initThumbnailGenerator(config.Content, extractor, log)
	if err != nil {
		return errors.Wrap(err, "initializing thumbnail generator")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	monitors := initFeedMonitors(ctx, config.FeedManager, service.ArticleRepo(), searchProvider, thumbnailer, log)
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

	articleProcessors, err := initArticleProcessors(config.Content.Article.Processors, config.Content.ProxyHTTPURLTemplate, log)
	if err != nil {
		return errors.WithMessage(err, "initializing article processors")
	}

	handler, err = api.Mux(ctx, service, feedManager, hubbub, searchProvider, extractor, fs, articleProcessors, config, log)
	if err != nil {
		return errors.WithMessage(err, "creating api mux")
	}

	mux.Handle("/api", handler)

	feedManager.Start(ctx)

	server := makeHTTPServer(mux)

	if serverDevelPort > 0 {
		server.Addr = fmt.Sprintf(":%d", serverDevelPort)

		if err = server.ListenAndServe(); err != nil {
			return errors.Wrap(err, "starting devel server")
		}
	}

	server.Addr = fmt.Sprintf("%s:%d", config.Server.Address, config.Server.Port)

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
		return nil, nil, errors.Wrapf(err, "opening bolt db", config.SessionStoragePath)
	}

	return db, boltstore.New(db, 30), nil
}

func initLog(config config.Log) readeef.Logger {
	return readeef.NewLogger(config)
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

func initArticleProcessors(names []string, proxyTemplate string, log readeef.Logger) ([]processor.Article, error) {
	var processors []processor.Article

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
			}
		case "insert-thumbnail-target":
			processors = append(processors, contentProcessor.NewInsertThumbnailTarget(log))
		}
	}

	return processors, nil
}

func initParserProcessors(names []string, proxyTemplate string, log readeef.Logger) ([]parser.Processor, error) {
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
			}
		case "cleanup":
			processors = append(processors, processor.NewCleanup(log))
		case "top-image-marker":
			processors = append(processors, processor.NewTopImageMarker(log))
		}
	}

	return processors, nil
}

func initSearchProvider(config config.Content, repo repo.Article, log readeef.Logger) search.Provider {
	var searchProvider search.Provider
	var err error

	switch config.Search.Provider {
	case "elastic":
		if searchProvider, err = search.NewElastic(
			config.Search.ElasticURL,
			config.Search.BatchSize,
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
			log,
		); err != nil {
			log.Printf("Error initializing Bleve search: %+v\n", err)
		}
	}

	if searchProvider != nil {
		if searchProvider.IsNewIndex() {
			go func() {
				if err := search.Reindex(searchProvider, repo); err != nil {
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
		if ce, err := extractor.NewReadability(config.Extract.ReadabilityKey); err == nil {
			return ce, nil
		} else {
			return nil, errors.WithMessage(err, "initializing Readability extract generator")
		}
	case "goose":
		fallthrough
	default:
		if ce, err := extractor.NewGoose("templates", fs); err == nil {
			return ce, nil
		} else {
			return nil, errors.WithMessage(err, "initializing Goose extract generator")
		}
	}
}

func initThumbnailGenerator(config config.Content, extract extract.Generator, log readeef.Logger) (thumbnail.Generator, error) {
	switch config.ThumbnailGenerator {
	case "extract":
		if t, err := thumbnail.FromExtract(extract, log); err == nil {
			return t, nil
		} else {
			return nil, errors.WithMessage(err, "initializing Extract thumbnail generator")
		}
	case "description":
		fallthrough
	default:
		return thumbnail.FromDescription(log), nil
	}
}

func initFeedMonitors(
	ctx context.Context,
	config config.FeedManager,
	repo repo.Article,
	searchProvider search.Provider,
	thumbnailer content.Thumbnailer,
	log readeef.Logger,
) []monitor.Feed {
	monitors := []monitor.Feed{monitor.NewUnread(ctx, repo, log)}

	for _, m := range config.Monitors {
		switch m {
		case "index":
			if searchProvider != nil {
				monitors = append(monitors, monitor.NewIndex(searchProvider, log))
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
	log readeef.Logger,
) (*readeef.Hubbub, error) {
	if config.Hubbub.CallbackURL != "" {
		hubbub := readeef.NewHubbub(config, log, "/api/v2/hubbub", feedManager)

		if err := hubbub.InitSubscriptions(service); err != nil {
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
