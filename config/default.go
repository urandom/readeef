package config

import (
	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

func defaultConfig() (Config, error) {
	var def Config

	err := toml.Unmarshal([]byte(defaultCfg), &def)

	if err != nil {
		return Config{}, errors.Wrap(err, "parsing default config")
	}

	def.API.Version = apiversion
	return def, nil
}

var defaultCfg = `
[server]
	port = 8080
	devel = true
[log]
	level = "error"    # error, info, debug
	file = "-"         # stderr, or a filename
	formatter = "text" # text, json
	access-file = "-"  # stdout or a filename
[api]
	emulators = []     # ["tt-rss", "fever"]
[db]
	driver = "sqlite3"
	connect = "file:./storage/content.sqlite3?cache=shared&mode=rwc"
[auth]
	session-storage-path = "./storage/session.db"
	token-storage-path = "./storage/token.db"
[feed-manager]
	update-interval = "30m"
	monitors = ["index", "thumbnailer"]
[timeout]
	connect = "1s"
	read-write = "2s"
[hubbub]
	from = "readeef"
[popularity]
	delay = "5s"
	providers = ["Facebook", "Reddit"]
[feed-parser]
	processors = ["cleanup", "top-image-marker", "absolutize-urls"]
	proxy-http-url-template = "/proxy?url={{ . }}"
[content]
	extractor = "goose" # readability
	thumbnailer = "description"
	search-provider = "bleve"

	article-processors = ["insert-thumbnail-target"]

	search-batch-size = 100

	bleve-path = "./storage/search.bleve"
	elastic-url = "http://localhost:9200"
	proxy-http-url-template = "/proxy?url={{ . }}"
`
