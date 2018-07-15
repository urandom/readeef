package main

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

type config struct {
	URL string
}

var (
	configPath = filepath.Join(CONFIG_DIR, "config.toml")
)

func readConfig() (config, error) {
	var c config

	if _, err := toml.DecodeFile(configPath, &c); err != nil {
		return c, errors.Wrap(err, "reading config file")
	}

	return c, nil
}

func writeConfig(c config) error {
	if err := os.MkdirAll(configPath, 0755); err != nil {
		return errors.Wrap(err, "creating config directory")
	}

	if err := os.Rename(configPath, configPath+".bkp"); err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, "backing up config file")
	}

	f, err := os.Create(configPath)
	if err != nil {
		return errors.Wrap(err, "creating config file")
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(c); err != nil {
		return errors.Wrap(err, "encoding config")
	}

	return nil
}
