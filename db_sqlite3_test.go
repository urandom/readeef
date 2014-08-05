// +build sqlite3,cgo

package readeef

import "os"

func init() {
	os.Remove(file)

	db = NewDB("sqlite3", "file::memory:")
	if err := db.Connect(); err != nil {
		panic(err)
	}
}
