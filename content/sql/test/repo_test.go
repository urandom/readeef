package test

import (
	"os"
	"testing"

	"github.com/urandom/readeef/content/sql/sqlite3"
	dbo "github.com/urandom/readeef/content/sql/db"
	"github.com/urandom/readeef/tests"
	"github.com/urandom/webfw"
)

func TestRepo(t *testing.T) {
	tests.CheckBool(t, false, repo.HasErr())
}

var (
	logger = webfw.NewStandardLogger(os.Stderr, "", 0)
	db     = dbo.New(logger)
	repo   = sqlite3.NewRepo(db, logger)
)
