package legacy

import (
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/config"

	"gopkg.in/gcfg.v1"
)

var apiversion = 1

type Config struct {
	Server struct {
		Address  string
		Port     int
		CertFile string `gcfg:"cert-file"`
		KeyFile  string `gcfg:"key-file"`
		Devel    bool
	}
	Renderer struct {
		Base string
		Dir  string
	}
	Dispatcher struct {
		Middleware []string
	}
	Static struct {
		Dir      string
		Expires  string
		Prefix   string
		Index    string
		FileList bool `gcfg:"file-list"`
	}
	Session struct {
		Dir             string
		Secret          string
		Cipher          string   // optional: 16, 24 or 32 bytes, base64 encoded
		MaxAge          string   `gcfg:"max-age"`
		CleanupInterval string   `gcfg:"cleanup-interval"`
		CleanupMaxAge   string   `gcfg:"cleanup-max-age"`
		IgnoreURLPrefix []string `gcfg:"ignore-url-prefix"`
	}
	I18n struct {
		Dir              string
		Languages        []string `gcfg:"language"`
		FallbackLanguage string   `gcfg:"fallback-language"`
		IgnoreURLPrefix  []string `gcfg:"ignore-url-prefix"`
	}
	Sitemap struct {
		LocPrefix        string `gcfg:"location-prefix"`
		RelativeLocation string `gcfg:"relative-location"`
	}
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
		Secret           string
		TokenStoragePath string   `gcfg:"token-storage-path"`
		IgnoreURLPrefix  []string `gcfg:"ignore-url-prefix"`
	}
	Hubbub struct {
		CallbackURL  string `gcfg:"callback-url"` // http://www.example.com
		RelativePath string `gcfg:"relative-path"`
		From         string
	}

	Popularity struct {
		Delay     string
		Providers []string

		TwitterConsumerKey       string `gcfg:"twitter-consumer-key"`
		TwitterConsumerSecret    string `gcfg:"twitter-consumer-secret"`
		TwitterAccessToken       string `gcfg:"twitter-access-token"`
		TwitterAccessTokenSecret string `gcfg:"twitter-access-token-secret"`

		Converted struct {
			Delay time.Duration
		}
	}

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

func ReadConfig(path string) (config.Config, bool, error) {
	exists := false
	def, err := defaultConfig()

	if err != nil {
		return config.Config{}, exists, err
	}

	c := def

	if path != "" {
		err = gcfg.ReadFileInto(&c, path)

		if !os.IsNotExist(errors.Cause(err)) {
			exists = true
		}

		if err != nil && !os.IsNotExist(err) {
			return config.Config{}, exists, err
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

	cfg, err := c.convertToV2()
	if err != nil {
		return config.Config{}, exists, errors.WithMessage(err, "converting to V2")
	}

	return cfg, exists, nil
}

func defaultConfig() (Config, error) {
	var def Config

	err := gcfg.ReadStringInto(&def, DefaultCfg)

	if err != nil {
		return Config{}, err
	}

	def.API.Version = apiversion
	return def, nil
}

func (c Config) convertToV2() (config.Config, error) {
	cfg, err := config.Read("")
	if err != nil {
		return config.Config{}, errors.WithMessage(err, "creating config")
	}

	cfg.Server.Address = c.Server.Address
	cfg.Server.Port = c.Server.Port
	cfg.Server.CertFile = c.Server.CertFile
	cfg.Server.KeyFile = c.Server.KeyFile

	cfg.Log.AccessFile = c.Logger.AccessFile
	cfg.Log.File = c.Logger.File
	cfg.Log.Formatter = c.Logger.Formatter
	cfg.Log.Level = c.Logger.Level

	cfg.API.Emulators = c.API.Emulators
	cfg.Timeout = config.Timeout(c.Timeout)
	cfg.DB = config.DB(c.DB)
	cfg.FeedParser = config.FeedParser(c.FeedParser)
	cfg.FeedManager = config.FeedManager(c.FeedManager)

	cfg.Content.Article.Processors = c.Content.ArticleProcessors
	cfg.Content.Article.ProxyHTTPURLTemplate = c.Content.ProxyHTTPURLTemplate
	cfg.Content.Search.Provider = c.Content.SearchProvider
	cfg.Content.Search.BatchSize = c.Content.SearchBatchSize
	cfg.Content.Search.BlevePath = c.Content.BlevePath
	cfg.Content.Search.ElasticURL = c.Content.ElasticURL
	cfg.Content.Extract.Generator = c.Content.Extractor
	cfg.Content.Extract.ReadabilityKey = c.Content.ReadabilityKey
	cfg.Content.ThumbnailGenerator = c.Content.Thumbnailer

	cfg.Auth.Secret = c.Auth.Secret

	cfg.Hubbub.CallbackURL = c.Hubbub.CallbackURL
	cfg.Hubbub.From = c.Hubbub.From

	cfg.Popularity.Delay = c.Popularity.Delay
	cfg.Popularity.Providers = c.Popularity.Providers
	cfg.Popularity.Converted.Delay = c.Popularity.Converted.Delay
	cfg.Popularity.Twitter.AccessToken = c.Popularity.TwitterAccessToken
	cfg.Popularity.Twitter.AccessTokenSecret = c.Popularity.TwitterAccessTokenSecret
	cfg.Popularity.Twitter.ConsumerKey = c.Popularity.TwitterConsumerKey
	cfg.Popularity.Twitter.ConsumerSecret = c.Popularity.TwitterConsumerSecret

	return cfg, nil
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
