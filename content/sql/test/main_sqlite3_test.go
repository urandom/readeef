// +build sqlite3

package test

import (
	"os"
	"testing"

	_ "github.com/urandom/readeef/content/sql/db/sqlite3"
)

func TestMain(m *testing.M) {
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

	ret := m.Run()

	os.Exit(ret)
}
