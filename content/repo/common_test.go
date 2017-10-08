package repo_test

import (
	"os"
	"testing"

	"github.com/urandom/readeef/config"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
)

var (
	skip = true

	cfg     config.Log
	logger  log.Log
	service repo.Service
)

func skipTest(t *testing.T) {
	if skip {
		t.Skip("No database tag selected")
	}
}

func init() {
	cfg.Converted.Writer = os.Stderr
	cfg.Converted.Prefix = "[testing] "

	logger = log.WithStd(cfg)
}
