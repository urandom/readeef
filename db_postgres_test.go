// +build postgres

package readeef

func init() {
	db = NewDB("postgres", "host=/var/run/postgresql user=urandom dbname=readeef-test sslmode=disable")
	if err := db.Connect(); err != nil {
		panic(err)
	}
}
