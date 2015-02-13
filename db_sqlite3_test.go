// +build sqlite3,cgo

package readeef

import (
	"os"

	"github.com/urandom/webfw"
)

func init() {
	file := "readeef-test.sqlite"
	conn := "file:./" + file + "?cache=shared&mode=rwc"

	os.Remove(file)

	db = NewDB("sqlite3", conn, webfw.NewStandardLogger(os.Stderr, "", 0))
	if err := db.Connect(); err != nil {
		panic(err)
	}
}
