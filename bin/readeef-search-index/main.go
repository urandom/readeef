package main

import (
	"flag"
	"fmt"
	"log"
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

	logger := log.New(os.Stderr, "", 0)
	readeef.InitDebug(logger, cfg)

	db := readeef.NewDB(cfg.DB.Driver, cfg.DB.Connect)
	if err := db.Connect(); err != nil {
		exitWithError(fmt.Sprintf("Error connecting to database: %v", err))
	}

	si, err := readeef.NewSearchIndex(cfg, db, logger)

	if err != nil {
		exitWithError(fmt.Sprintf("Error creating search index: %v", err))
	}

	readeef.Debug.Println("Getting all articles")

	if articles, err := db.GetAllArticles(); err == nil {
		for _, a := range articles {
			readeef.Debug.Printf("Indexing article '%s' from feed id '%d'\n", a.Id, a.FeedId)
			if err := si.Index(a); err != nil {
				logger.Printf("Error indexing article %s from feed %d: %v\n", a.Id, a.FeedId, err)
			}
		}
	} else {
		exitWithError(fmt.Sprintf("Error getting all articles: %v", err))
	}
}

func exitWithError(err string) {
	fmt.Fprintf(os.Stderr, err+"\n")
	os.Exit(1)
}
