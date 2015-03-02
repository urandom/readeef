// +build postgres

package test

import (
	"os"
	"testing"

	_ "github.com/urandom/readeef/db/postgres"
)

func TestMain(m *testing.M) {
	if err := db.Open("postgres", "host=/var/run/postgresql user=urandom dbname=readeef-test sslmode=disable"); err != nil {
		panic(err)
	}
	ret := m.Run()

	db.Exec("TRUNCATE articles CASCADE")
	db.Exec("TRUNCATE articles_scores CASCADE")
	db.Exec("TRUNCATE feed_images CASCADE")
	db.Exec("TRUNCATE feeds CASCADE")
	db.Exec("TRUNCATE hubbub_subscriptions CASCADE")
	db.Exec("TRUNCATE users CASCADE")
	db.Exec("TRUNCATE users_articles_fav CASCADE")
	db.Exec("TRUNCATE users_articles_read CASCADE")
	db.Exec("TRUNCATE users_feeds CASCADE")
	db.Exec("TRUNCATE users_feeds_tags CASCADE")

	os.Exit(ret)
}
