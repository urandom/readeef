package main

import (
	"flag"

	"github.com/pkg/errors"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/config"
	"github.com/urandom/readeef/content/repo/sql"
	"github.com/urandom/readeef/content/search"
)

var (
	searchIndexVerbose bool
)

func runSearchIndex(config config.Config, args []string) error {
	if searchIndexVerbose {
		config.Log.Level = "debug"
	}

	log := readeef.NewLogger(config.Log)
	service, err := sql.NewService(config.DB.Driver, config.DB.Connect, log)
	if err != nil {
		return errors.WithMessage(err, "creating content service")
	}

	searchProvider := initSearchProvider(config.Content, repo, log)
	if searchProvider == nil {
		return errors.Errorf("unknown search provider %s", config.Content.SearchProvider)
	}

	log.Info("Starting feed indexing")

	if err := search.Reindex(searchProvider, service.ArticleRepo()); err != nil {
		return errors.WithMessage(err, "indexing all feeds")
	}

	return nil
}

func init() {
	flags := flag.NewFlagSet("search-index", flag.ExitOnError)
	flags.BoolVar(&searchIndexVerbose, "verbose", false, "verbose output")

	commands = append(commands, Command{
		Name:  "search-index",
		Desc:  "re-index all feeds",
		Flags: flags,
		Run:   runSearchIndex,
	})
}
