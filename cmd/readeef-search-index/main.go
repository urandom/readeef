package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/urandom/readeef"
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

	db := readeef.NewDB(cfg.DB.Driver, cfg.DB.Connect, logger)
	if err := db.Connect(); err != nil {
		exitWithError(fmt.Sprintf("Error connecting to database: %v", err))
	}

	si, err := readeef.NewSearchIndex(cfg, db, logger)

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
