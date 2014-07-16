package readeef

import (
	"os"

	"code.google.com/p/gcfg"
)

type Config struct {
	DB struct {
		Driver  string
		Connect string
	}
	Auth struct {
		Secret          string
		IgnoreURLPrefix []string `gcfg:"ignore-url-prefix"`
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

	return c, nil
}

func defaultConfig() (Config, error) {
	var def Config

	err := gcfg.ReadStringInto(&def, cfg)

	if err != nil {
		return Config{}, err
	}

	return def, nil
}

var cfg string = `
[db]
	driver = sqlite3
	connect = file:./readeef.sqlite3?cache=shared&mode=rwc
`
