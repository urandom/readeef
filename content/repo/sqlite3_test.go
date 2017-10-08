// +build sqlite3

package repo_test

import (
	"os"
	"testing"

	"github.com/urandom/readeef/content/repo/sql"
	"github.com/urandom/readeef/content/repo/sql/db"
	_ "github.com/urandom/readeef/content/repo/sql/db/sqlite3"
)

func TestMain(m *testing.M) {
	db := db.New(logger)
	if err := db.Open("sqlite3", "file:/tmp/readeef-test.sqlite3?cache=shared"); err != nil {
		// if err := db.Open("sqlite3", "file::memory:?cache=shared"); err != nil {
		panic(err)
	}

	db.Exec("DELETE FROM articles")
	db.Exec("DELETE FROM articles_scores")
	db.Exec("DELETE FROM feed_images")
	db.Exec("DELETE FROM feeds")
	db.Exec("DELETE FROM hubbub_subscriptions")
	db.Exec("DELETE FROM users")
	db.Exec("DELETE FROM users_articles_states")
	db.Exec("DELETE FROM users_feeds")
	db.Exec("DELETE FROM users_feeds_tags")

	var err error
	service, err = sql.NewService("sqlite3", "file:/tmp/readeef-test.sqlite3?cache=shared", logger)
	if err != nil {
		panic(err)
	}

	skip = false
	ret := m.Run()

	os.Exit(ret)
}
