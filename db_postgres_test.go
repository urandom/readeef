// +build postgres

package readeef

import (
	"os"

	"github.com/urandom/webfw"
)

func init() {
	db = NewDB("postgres", "host=/var/run/postgresql user=urandom dbname=readeef-test sslmode=disable", webfw.NewStandardLogger(os.Stderr, "", 0))
	if err := db.Connect(); err != nil {
		panic(err)
	}
}
