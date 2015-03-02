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

	db.Exec("TRUNCATE users CASCADE")

	os.Exit(ret)
}
