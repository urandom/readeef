package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base/search"
	"github.com/urandom/readeef/content/repo"
	_ "github.com/urandom/readeef/content/sql/db/postgres"
)

func main() {
	confpath := flag.String("config", "", "readeef config path")

	flag.Parse()

	cfg, err := readeef.ReadConfig(*confpath)
	if err != nil {
		exitWithError(fmt.Sprintf("Error reading config from path '%s': %v", *confpath, err))
	}

	logger := readeef.NewLogger(cfg)
	repo, err := repo.New(cfg.DB.Driver, cfg.DB.Connect, logger)
	if err != nil {
		exitWithError(fmt.Sprintf("Error connecting to database: %v", err))
	}

	var sp content.SearchProvider

	switch cfg.Content.SearchProvider {
	case "elastic":
		if sp, err = search.NewElastic(cfg.Content.ElasticURL, cfg.Content.SearchBatchSize, logger); err != nil {
			exitWithError(fmt.Sprintf("Error initializing Elastic search: %v\n", err))
		}
	case "bleve":
		fallthrough
	default:
		if sp, err = search.NewBleve(cfg.Content.BlevePath, cfg.Content.SearchBatchSize, logger); err != nil {
			exitWithError(fmt.Sprintf("Error initializing Bleve search: %v\n", err))
		}
	}

	logger.Infoln("Getting all articles")

	if err := sp.IndexAllFeeds(repo); err != nil {
		exitWithError(fmt.Sprintf("Error indexing all articles: %v", err))
	}
}

func exitWithError(err string) {
	fmt.Fprintf(os.Stderr, err+"\n")
	os.Exit(1)
}
