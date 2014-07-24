package readeef

import (
	"os"
	"time"

	"code.google.com/p/gcfg"
)

var apiversion = 1

type Config struct {
	API struct {
		Version int
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
		CallbackURL  string `gcfg:"callback-url"` // http://www.example.com/dispatcher-path
		RelativePath string `gcfg:"relative-path"`
		From         string
	}
}

func ReadConfig(path ...string) (Config, error) {
	def, err := defaultConfig()

	if err != nil {
		return Config{}, err
	}

	if len(path) == 0 {
		return def, nil
	}

	c := def

	err = gcfg.ReadFileInto(&c, path[0])

	if err != nil {
		if os.IsNotExist(err) {
			return def, nil
		}

		return Config{}, err
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
[db]
	driver = sqlite3
	connect = file:./readeef.sqlite3?cache=shared&mode=rwc
[timeout]
	connect = 1s
	read-write = 2s
[hubbub]
	relative-path = hubbub
	from = readeef
`
