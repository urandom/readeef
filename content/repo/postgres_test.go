// +build postgres

package repo_test

import (
	"os"
	"testing"

	"github.com/urandom/readeef/content/repo/sql"
	"github.com/urandom/readeef/content/repo/sql/db"
	_ "github.com/urandom/readeef/content/repo/sql/db/postgres"
)

func TestMain(m *testing.M) {
	db := db.New(logger)
	if err := db.Open("postgres", "host=/var/run/postgresql user=urandom dbname=readeef-test sslmode=disable"); err != nil {
		panic(err)
	}

	db.Exec("TRUNCATE articles CASCADE")
	db.Exec("TRUNCATE articles_scores CASCADE")
	db.Exec("TRUNCATE feed_images CASCADE")
	db.Exec("TRUNCATE feeds CASCADE")
	db.Exec("TRUNCATE hubbub_subscriptions CASCADE")
	db.Exec("TRUNCATE users CASCADE")
	db.Exec("TRUNCATE users_articles_states CASCADE")
	db.Exec("TRUNCATE users_feeds CASCADE")
	db.Exec("TRUNCATE users_feeds_tags CASCADE")

	db.Close()

	var err error
	service, err = sql.NewService("postgres", "host=/var/run/postgresql user=urandom dbname=readeef-test sslmode=disable", logger)
	if err != nil {
		panic(err)
	}

	skip = false
	ret := m.Run()

	os.Exit(ret)
}
