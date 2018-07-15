package main

import (
	"os"
	"path/filepath"
)

var (
	OPEN       = []string{"cmd", "/c", "start"}
	CONFIG_DIR = filepath.Join(os.Getenv("%APPDATA%"), "readeef-client")
)
