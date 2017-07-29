package config

import "time"

const apiversion = 2

type Server struct {
	Address  string `toml:"address"`
	Port     int    `toml:"port"`
	CertFile string `toml:"cert-file"`
	KeyFile  string `toml:"key-file"`
	Devel    bool   `toml:"devel"`
}

type Log struct {
	Level      string `toml:"level"`
	File       string `toml:"file"`
	AccessFile string `toml:"access-file"`
	Formatter  string `toml:"formatter"`
}

type API struct {
	Version   int      `toml:"version"`
	Emulators []string `toml:"emulators"`
}

type Timeout struct {
	Connect   string `toml:"connect"`
	ReadWrite string `toml:"read-write"`

	Converted struct {
		Connect   time.Duration
		ReadWrite time.Duration
	}
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

	Converted struct {
		Delay time.Duration
	}
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
	}
}

type Content struct {
	Extractor         string   `toml:"extractor"`
	Thumbnailer       string   `toml:"thumbnailer"`
	SearchProvider    string   `toml:"search-provider"`
	ArticleProcessors []string `toml:"article-processors"`

	SearchBatchSize int64 `toml:"search-batch-size"`

	ReadabilityKey       string `toml:"readability-key"`
	BlevePath            string `toml:"bleve-path"`
	ElasticURL           string `toml:"elastic-url"`
	ProxyHTTPURLTemplate string `toml:"proxy-http-url-template"`
}

type converter interface {
	convert()
}

func (c *API) convert() {
	c.Version = apiversion
}

func (c *Timeout) convert() {
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

func (c *Popularity) convert() {
	if d, err := time.ParseDuration(c.Delay); err == nil {
		c.Converted.Delay = d
	} else {
		c.Converted.Delay = 5 * time.Second
	}
}

func (c *FeedManager) convert() {
	if d, err := time.ParseDuration(c.UpdateInterval); err == nil {
		c.Converted.UpdateInterval = d
	} else {
		c.Converted.UpdateInterval = 30 * time.Minute
	}
}
