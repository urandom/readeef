package config

import (
	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

func defaultConfig() (Config, error) {
	var def Config

	err := toml.Unmarshal([]byte(DefaultCfg), &def)

	if err != nil {
		return Config{}, errors.Wrap(err, "parsing default config")
	}

	def.API.Version = apiversion
	return def, nil
}

// DefaultCfg shows the default configuration of the readeef server
var DefaultCfg = `
[server]
	port = 8080
[server.auto-cert]
	storage-path = "./storage/certs"
[log]
	level = "info"     # error, info, debug
	file = "-"         # stderr, or a filename
	formatter = "text" # text, json
	access-file = ""   # stdout or a filename
[api]
	emulators = []     # ["tt-rss", "fever"]
[api.limits]
	articles-per-query = 200
[db]
	driver = "sqlite3"
	connect = "file:./storage/content.sqlite3?cache=shared&mode=rwc&_busy_timeout=50000000"
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
	# providers = ["Reddit", "Twitter"]
[feed-parser]
	processors = ["cleanup", "top-image-marker", "absolutize-urls", "unescape"]
	proxy-http-url-template = "/proxy?url={{ . }}"
[content]
	thumbnail-generator = "description"
[content.extract]
	generator = "goose" # readability
[content.search]
	provider = "bleve"
	batch-size = 100
	bleve-path = "./storage/search.bleve"
	elastic-url = "http://localhost:9200"
[content.article]
	processors = ["insert-thumbnail-target"]
	proxy-http-url-template = "/proxy?url={{ . }}"
[ui]
	path = "./rf-ng/ui"
`
