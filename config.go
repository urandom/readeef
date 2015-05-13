package readeef

import (
	"os"
	"time"

	"code.google.com/p/gcfg"
)

var apiversion = 1

type Config struct {
	Logger struct {
		Level      string
		File       string
		AccessFile string `gcfg:"access-file"`
		Formatter  string
	}
	API struct {
		Version int
		Fever   bool
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
	Updater struct {
		Interval string

		Converted struct {
			Interval time.Duration
		}
	}
	ArticleFormatter struct {
		ReadabilityKey                 string `gcfg:"readability-key"`
		ConvertLinksToProtocolRelative bool   `gcfg:"convert-links-to-protocol-relative"`
	} `gcfg:"article-formatter"`

	SearchIndex struct {
		BlevePath string `gcfg:"bleve-path"`
		BatchSize int64  `gcfg:"batch-size"`
	} `gcfg:"search-index"`

	Popularity struct {
		Delay     string
		Providers []string

		Converted struct {
			Delay time.Duration
		}
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

	if d, err := time.ParseDuration(c.Updater.Interval); err == nil {
		c.Updater.Converted.Interval = d
	} else {
		c.Updater.Converted.Interval = 30 * time.Minute
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

	err := gcfg.ReadStringInto(&def, cfg)

	if err != nil {
		return Config{}, err
	}

	def.API.Version = apiversion
	return def, nil
}

var cfg string = `
[logger]
	level = error # error, info, debug
	file = - # stderr, or a filename
	formatter = text # text, json
	access-file = - # stdout or a filename
[db]
	driver = sqlite3
	connect = file:./readeef.sqlite3?cache=shared&mode=rwc
[timeout]
	connect = 1s
	read-write = 2s
[hubbub]
	relative-path = /hubbub
	from = readeef
[search-index]
	bleve-path = ./readeef.bleve
	batch-size = 100
[popularity]
	delay = 5s
	providers
	providers = Facebook
	providers = GoogleP
	providers = Twitter
	providers = Reddit
	providers = Linkedin
	providers = StumbleUpon
`
