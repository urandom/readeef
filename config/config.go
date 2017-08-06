package config

import (
	"io/ioutil"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

// Config is the readeef configuration
type Config struct {
	Server      Server      `toml:"server"`
	Log         Log         `toml:"log"`
	API         API         `toml:"api"`
	Timeout     Timeout     `toml:"timeout"`
	DB          DB          `toml:"db"`
	Auth        Auth        `toml:"auth"`
	Hubbub      Hubbub      `toml:"hubbub"`
	Popularity  Popularity  `toml:"popularity"`
	FeedParser  FeedParser  `toml:"feed-parser"`
	FeedManager FeedManager `toml:"feed-manager"`
	Content     Content     `toml:"content"`
}

// Read loads the config data from the given path
func Read(path string) (Config, error) {
	c, err := defaultConfig()

	if err != nil {
		return Config{}, errors.WithMessage(err, "initializing default config")
	}

	if path != "" {
		b, err := ioutil.ReadFile(path)
		if err != nil {
			return Config{}, errors.Wrapf(err, "reading config from %s", path)
		}

		if err = toml.Unmarshal(b, &c); err != nil {
			return Config{}, errors.Wrapf(err, "unmarshaling toml config from %s", path)
		}
	}

	for _, c := range []converter{&c.API, &c.Timeout, &c.FeedManager, &c.Popularity} {
		c.convert()
	}

	return c, nil

}
