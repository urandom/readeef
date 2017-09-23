// +build sqlite3

package sql

import (
	"os"
	"testing"

	_ "github.com/urandom/readeef/content/repo/sql/db/sqlite3"
)

func TestMain(m *testing.M) {
	if err := dbo.Open("sqlite3", "file:/tmp/readeef-test.sqlite3?cache=shared"); err != nil {
		// if err := db.Open("sqlite3", "file::memory:?cache=shared"); err != nil {
		panic(err)
	}

	dbo.Exec("DELETE FROM articles")
	dbo.Exec("DELETE FROM articles_scores")
	dbo.Exec("DELETE FROM feed_images")
	dbo.Exec("DELETE FROM feeds")
	dbo.Exec("DELETE FROM hubbub_subscriptions")
	dbo.Exec("DELETE FROM users")
	dbo.Exec("DELETE FROM users_articles_states")
	dbo.Exec("DELETE FROM users_feeds")
	dbo.Exec("DELETE FROM users_feeds_tags")

	skip = false
	ret := m.Run()

	os.Exit(ret)
}
