package test

import (
	"os"
	"testing"

	"github.com/urandom/readeef"
	dbo "github.com/urandom/readeef/content/sql/db"
	"github.com/urandom/readeef/content/sql/sqlite3"
	"github.com/urandom/readeef/tests"
)

func TestRepo(t *testing.T) {
	tests.CheckBool(t, false, repo.HasErr())
}

var (
	logger = readeef.NewStandardLogger(os.Stderr, "", 0)
	db     = dbo.New(logger)
	repo   = sqlite3.NewRepo(db, logger)
)
