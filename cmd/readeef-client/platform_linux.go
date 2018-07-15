package main

import (
	"os"
	"path/filepath"
)

var (
	OPEN       = []string{"xdg-open"}
	CONFIG_DIR = filepath.Join(os.Getenv("HOME"), ".config/readeef-client")
)
