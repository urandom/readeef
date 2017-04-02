package readeef

import (
	"os"
	"time"

	"github.com/urandom/readeef/popularity"
	"github.com/urandom/webfw"

	"gopkg.in/gcfg.v1"
)

var apiversion = 1

type Config struct {
	webfw.Config
	Logger struct {
		Level      string
		File       string
		AccessFile string `gcfg:"access-file"`
		Formatter  string
	}
	API struct {
		Version   int
		Emulators []string
	}
	Timeout struct {
		Connect   string
		ReadWrite string `gcfg:"read-write"`

		Converted struct {
			Connect   time.Duration
			ReadWrite time.Duration
		}
	}
	DB struct {
		Driver  string
		Connect string
	}
	Auth struct {
		Secret          string
		IgnoreURLPrefix []string `gcfg:"ignore-url-prefix"`
	}
	Hubbub struct {
		CallbackURL  string `gcfg:"callback-url"` // http://www.example.com
		RelativePath string `gcfg:"relative-path"`
		From         string
	}

	Popularity popularity.Config

	FeedParser struct {
		Processors []string

		ProxyHTTPURLTemplate string `gcfg:"proxy-http-url-template"`
	} `gcfg:"feed-parser"`

	FeedManager struct {
		UpdateInterval string `gcfg:"update-interval"`

		Monitors []string

		Converted struct {
			UpdateInterval time.Duration
		}
	} `gcfg:"feed-manager"`

	Content struct {
		Extractor         string
		Thumbnailer       string
		SearchProvider    string   `gcfg:"search-provider"`
		ArticleProcessors []string `gcfg:"article-processors"`

		SearchBatchSize int64 `gcfg:"search-batch-size"`

		ReadabilityKey       string `gcfg:"readability-key"`
		BlevePath            string `gcfg:"bleve-path"`
		ElasticURL           string `gcfg:"elastic-url"`
		ProxyHTTPURLTemplate string `gcfg:"proxy-http-url-template"`
	}
}

func ReadConfig(path ...string) (Config, error) {
	def, err := defaultConfig()

	if err != nil {
		return Config{}, err
	}

	c := def

	if len(path) != 0 {
		err = gcfg.ReadFileInto(&c, path[0])

		if err != nil && !os.IsNotExist(err) {
			return Config{}, err
		}
	}

	c.API.Version = apiversion

	if d, err := time.ParseDuration(c.Timeout.Connect); err == nil {
		c.Timeout.Converted.Connect = d
	} else {
		c.Timeout.Converted.Connect = time.Second
	}

	if d, err := time.ParseDuration(c.Timeout.ReadWrite); err == nil {
		c.Timeout.Converted.ReadWrite = d
	} else {
		c.Timeout.Converted.ReadWrite = time.Second
	}

	if d, err := time.ParseDuration(c.FeedManager.UpdateInterval); err == nil {
		c.FeedManager.Converted.UpdateInterval = d
	} else {
		c.FeedManager.Converted.UpdateInterval = 30 * time.Minute
	}

	if d, err := time.ParseDuration(c.Popularity.Delay); err == nil {
		c.Popularity.Converted.Delay = d
	} else {
		c.Popularity.Converted.Delay = 5 * time.Second
	}

	return c, nil
}

func defaultConfig() (Config, error) {
	var def Config

	err := gcfg.ReadStringInto(&def, webfw.DefaultCfg)
	if err == nil {
		err = gcfg.ReadStringInto(&def, DefaultCfg)
	}

	if err != nil {
		return Config{}, err
	}

	def.API.Version = apiversion
	return def, nil
}

// Default configuration
var DefaultCfg string = `
[logger]
	level = error # error, info, debug
	file = - # stderr, or a filename
	formatter = text # text, json
	access-file = - # stdout or a filename
[api]
	emulators
	# emulators = tt-rss
	# emulators = fever
[db]
	driver = sqlite3
	connect = file:./readeef.sqlite3?cache=shared&mode=rwc
[feed-manager]
	update-interval = 30m
	monitors
	monitors = index
	monitors = thumbnailer
[timeout]
	connect = 1s
	read-write = 2s
[hubbub]
	relative-path = /hubbub
	from = readeef
[popularity]
	delay = 5s
	providers
	providers = Facebook
	providers = GoogleP
	providers = Reddit
	providers = Linkedin
	providers = StumbleUpon
[feed-parser]
	processors
	processors = cleanup
	processors = top-image-marker
	processors = absolutize-urls
	# processors = relative-url
	# processors = proxy-http

	proxy-http-url-template = "/proxy?url={{ . }}"
[content]
	extractor = goose # readability
	thumbnailer = description
	search-provider = bleve

	article-processors
	article-processors = insert-thumbnail-target
	# article-processors = relative-url
	# article-processors = proxy-http

	search-batch-size = 100

	bleve-path = ./readeef.bleve
	elastic-url = http://localhost:9200
	proxy-http-url-template = "/proxy?url={{ . }}"
`
