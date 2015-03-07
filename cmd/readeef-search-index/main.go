package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/urandom/readeef"
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

	if cfg.SearchIndex.BlevePath == "" {
		exitWithError("No bleve-path in search-index section of the config")
	}

	logger := readeef.NewLogger(cfg)
	repo, err := repo.New(cfg.DB.Driver, cfg.DB.Connect, logger)
	if err != nil {
		exitWithError(fmt.Sprintf("Error connecting to database: %v", err))
	}

	si, err := readeef.NewSearchIndex(repo, cfg, logger)

	if err != nil {
		exitWithError(fmt.Sprintf("Error creating search index: %v", err))
	}

	logger.Infoln("Getting all articles")

	if err := si.IndexAllArticles(); err != nil {
		exitWithError(fmt.Sprintf("Error indexing all articles: %v", err))
	}
}

func exitWithError(err string) {
	fmt.Fprintf(os.Stderr, err+"\n")
	os.Exit(1)
}
