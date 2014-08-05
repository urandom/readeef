// +build sqlite3,cgo

package readeef

import "os"

func init() {
	file := "readeef-test.sqlite"
	//conn := "file:./" + file + "?cache=shared&mode=rwc"

	os.Remove(file)

	db = NewDB("sqlite3", "file::memory:")
	if err := db.Connect(); err != nil {
		panic(err)
	}
}
