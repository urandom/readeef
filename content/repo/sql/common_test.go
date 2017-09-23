package sql

import (
	"os"
	"testing"

	"github.com/urandom/readeef/content/repo/sql/db"
	"github.com/urandom/readeef/log"
)

var (
	skip    = true
	logger  = log.WithStd(os.Stderr, "testing", 0)
	dbo     = db.New(logger)
	service = Service{dbo, logger}
)

func skipTest(t *testing.T) {
	if skip {
		t.Skip("No database tag selected")
	}
}
