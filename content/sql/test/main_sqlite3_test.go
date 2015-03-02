// +build sqlite3

package test

import (
	"os"
	"testing"

	_ "github.com/urandom/readeef/db/sqlite3"
)

func TestMain(m *testing.M) {
	if err := db.Open("sqlite3", "file::memory:?cache=shared"); err != nil {
		panic(err)
	}
	ret := m.Run()

	os.Exit(ret)
}
