package sql

import (
	"os"
	"testing"

	"github.com/urandom/readeef/config"
	"github.com/urandom/readeef/content/repo/sql/db"
	"github.com/urandom/readeef/log"
)

var (
	skip = true

	cfg     config.Log
	logger  log.Log
	dbo     *db.DB
	service Service
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
	dbo = db.New(logger)
	service = Service{dbo, logger}
}
