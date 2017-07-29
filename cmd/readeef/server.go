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
	"github.com/urandom/readeef/content/base/extractor"
	"github.com/urandom/readeef/content/base/monitor"
	contentProcessor "github.com/urandom/readeef/content/base/processor"
	"github.com/urandom/readeef/content/base/search"
	"github.com/urandom/readeef/content/base/thumbnailer"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/content/repo"
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

	repo, err := repo.New(config.DB.Driver, config.DB.Connect, log)
	if err != nil {
		return errors.WithMessage(err, "creating content repo")
	}

	if processors, err := initArticleProcessors(config.Content.ArticleProcessors, config.Content.ProxyHTTPURLTemplate, log); err == nil {
		repo.ArticleProcessors(processors)
	} else {
		return errors.WithMessage(err, "initializing article processors")
	}

	if err = initAdminUser(repo, []byte(config.Auth.Secret)); err != nil {
		return errors.WithMessage(err, "initializing admin user")
	}

	feedManager := readeef.NewFeedManager(repo, config, log)

	if processors, err := initParserProcessors(config.FeedParser.Processors, config.FeedParser.ProxyHTTPURLTemplate, log); err == nil {
		for _, p := range processors {
			feedManager.AddParserProcessor(p)
		}
	} else {
		return errors.WithMessage(err, "initializing parser processors")
	}

	searchProvider := initSearchProvider(config.Content, repo, log)

	extractor, err := initContentExtractor(config.Content, fs)
	if err != nil {
		return errors.WithMessage(err, "initializing content extractor")
	}

	thumbnailer, err := initThumbnailer(config.Content, extractor, log)
	if err != nil {
		return errors.Wrap(err, "initializing thumbnailer")
	}

	monitors := initFeedMonitors(config.FeedManager, repo, searchProvider, thumbnailer, log)
	for _, m := range monitors {
		feedManager.AddFeedMonitor(m)
	}

	hubbub, err := initHubbub(config, repo, monitors, feedManager, log)
	if err != nil {
		return errors.WithMessage(err, "initializing hubbub")
	}

	if hubbub != nil {
		feedManager.SetHubbub(hubbub)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler, err = api.Mux(ctx, repo, feedManager, hubbub, searchProvider, extractor, fs, config, log)
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

func initArticleProcessors(names []string, proxyTemplate string, log readeef.Logger) ([]content.ArticleProcessor, error) {
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
			log.Printf("Error initializing Elastic search: %+v\n", err)
		}
	case "bleve":
		fallthrough
	default:
		if searchProvider, err = search.NewBleve(
			config.BlevePath,
			config.SearchBatchSize,
			log,
		); err != nil {
			log.Printf("Error initializing Bleve search: %+v\n", err)
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

func initContentExtractor(config config.Content, fs http.FileSystem) (content.Extractor, error) {
	switch config.Extractor {
	case "readability":
		if ce, err := extractor.NewReadability(config.ReadabilityKey); err == nil {
			return ce, nil
		} else {
			return nil, errors.WithMessage(err, "initializing Readability extractor")
		}
	case "goose":
		fallthrough
	default:
		if ce, err := extractor.NewGoose("templates", fs); err == nil {
			return ce, nil
		} else {
			return nil, errors.WithMessage(err, "initializing Goose extractor")
		}
	}
}

func initThumbnailer(config config.Content, ce content.Extractor, log readeef.Logger) (content.Thumbnailer, error) {
	switch config.Thumbnailer {
	case "extract":
		if t, err := thumbnailer.NewExtract(ce, log); err == nil {
			return t, nil
		} else {
			return nil, errors.WithMessage(err, "initializing Extract thumbnailer")
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
	log readeef.Logger,
) []content.FeedMonitor {
	monitors := []content.FeedMonitor{monitor.NewUnread(repo, log)}

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
	repo content.Repo,
	monitors []content.FeedMonitor,
	feedManager *readeef.FeedManager,
	log readeef.Logger,
) (*readeef.Hubbub, error) {
	if config.Hubbub.CallbackURL != "" {
		hubbub := readeef.NewHubbub(config, log, "/api/v2/hubbub", feedManager)

		if err := hubbub.InitSubscriptions(repo); err != nil {
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
