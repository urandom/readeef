package config

import (
	"io"
	"os"
	"time"

	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

const apiversion = 2

type Server struct {
	Address  string `toml:"address"`
	Port     int    `toml:"port"`
	CertFile string `toml:"cert-file"`
	KeyFile  string `toml:"key-file"`

	AutoCert struct {
		Host        string `toml:"host"`
		StoragePath string `toml:"storage-path"`
	} `toml:"auto-cert"`
}

type Log struct {
	Level            string `toml:"level"`
	File             string `toml:"file"`
	AccessFile       string `toml:"access-file"`
	Formatter        string `toml:"formatter"`
	RepoCallDuration bool   `toml:"repo-call-duration"`

	Converted struct {
		Writer io.Writer
		Prefix string
	} `toml:"-"`
}

type API struct {
	Version   int      `toml:"version"`
	Emulators []string `toml:"emulators"`
	Limits    struct {
		ArticlesPerQuery int `toml:"articles-per-query"`
	} `toml:"limits"`
}

type Timeout struct {
	Connect   string `toml:"connect"`
	ReadWrite string `toml:"read-write"`

	Converted struct {
		Connect   time.Duration
		ReadWrite time.Duration
	} `toml:"-"`
}

type DB struct {
	Driver  string `toml:"driver"`
	Connect string `toml:"connect"`
}

type Auth struct {
	Secret             string `toml:"secret"`
	SessionStoragePath string `toml:"session-storage-path"`
	TokenStoragePath   string `toml:"token-storage-path"`
}

type Hubbub struct {
	CallbackURL string `toml:"callback-url"` // http://www.example.com
	From        string `toml:"from"`
}

type Popularity struct {
	Delay     string   `toml:"delay"`
	Providers []string `toml:"providers"`

	Twitter struct {
		ConsumerKey       string `toml:"consumer-key"`
		ConsumerSecret    string `toml:"consumer-secret"`
		AccessToken       string `toml:"access-token"`
		AccessTokenSecret string `toml:"access-token-secret"`
	} `toml:"twitter"`

	Reddit struct {
		ID       string `toml:"id"`
		Secret   string `toml:"secret"`
		Username string `toml:"username"`
		Password string `toml:"password"`
	} `toml:"reddit"`

	Converted struct {
		Delay time.Duration
	} `toml:"-"`
}

type FeedParser struct {
	Processors []string `toml:"processors"`

	ProxyHTTPURLTemplate string `toml:"proxy-http-url-template"`
}

type FeedManager struct {
	UpdateInterval string `toml:"update-interval"`

	Monitors []string `toml:"monitors"`

	Converted struct {
		UpdateInterval time.Duration
	} `toml:"-"`
}

type Content struct {
	ThumbnailGenerator string `toml:"thumbnail-generator"`

	Extract struct {
		Generator      string `toml:"generator"`
		ReadabilityKey string `toml:"readability-key"`
	} `toml:"extract"`

	Search struct {
		Provider   string `toml:"provider"`
		BatchSize  int64  `toml:"batch-size"`
		BlevePath  string `toml:"bleve-path"`
		ElasticURL string `toml:"elastic-url"`
	} `toml:"search"`

	Article struct {
		Processors           []string `toml:"processors"`
		ProxyHTTPURLTemplate string   `toml:"proxy-http-url-template"`
	} `toml:"article"`
}

type UI struct {
	Path string `toml:"path"`
}

type converter interface {
	Convert()
}

func (c *API) Convert() {
	c.Version = apiversion
}

func (c *Log) Convert() {
	if c.File == "-" {
		c.Converted.Writer = os.Stderr
	} else {
		c.Converted.Writer = &lumberjack.Logger{
			Filename:   c.File,
			MaxSize:    20,
			MaxBackups: 5,
			MaxAge:     28,
		}
	}
}

func (c *Timeout) Convert() {
	if d, err := time.ParseDuration(c.Connect); err == nil {
		c.Converted.Connect = d
	} else {
		c.Converted.Connect = time.Second
	}

	if d, err := time.ParseDuration(c.ReadWrite); err == nil {
		c.Converted.ReadWrite = d
	} else {
		c.Converted.ReadWrite = time.Second
	}
}

func (c *Popularity) Convert() {
	if d, err := time.ParseDuration(c.Delay); err == nil {
		c.Converted.Delay = d
	} else {
		c.Converted.Delay = 5 * time.Second
	}
}

func (c *FeedManager) Convert() {
	if d, err := time.ParseDuration(c.UpdateInterval); err == nil {
		c.Converted.UpdateInterval = d
	} else {
		c.Converted.UpdateInterval = 30 * time.Minute
	}
}
