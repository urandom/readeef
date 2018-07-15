package main

import (
	"os"
	"path/filepath"
)

var (
	OPEN       = []string{"open"}
	CONFIG_DIR = filepath.Join(os.Getenv("HOME"), "Library/Preferences/readeef-client")
)
