// +build postgres

package sql

import (
	"os"
	"testing"

	_ "github.com/urandom/readeef/content/repo/sql/db/postgres"
)

func TestMain(m *testing.M) {
	if err := dbo.Open("postgres", "host=/var/run/postgresql user=urandom dbname=readeef-test sslmode=disable"); err != nil {
		panic(err)
	}

	dbo.Exec("TRUNCATE articles CASCADE")
	dbo.Exec("TRUNCATE articles_scores CASCADE")
	dbo.Exec("TRUNCATE feed_images CASCADE")
	dbo.Exec("TRUNCATE feeds CASCADE")
	dbo.Exec("TRUNCATE hubbub_subscriptions CASCADE")
	dbo.Exec("TRUNCATE users CASCADE")
	dbo.Exec("TRUNCATE users_articles_states CASCADE")
	dbo.Exec("TRUNCATE users_feeds CASCADE")
	dbo.Exec("TRUNCATE users_feeds_tags CASCADE")

	skip = false
	ret := m.Run()

	os.Exit(ret)
}
