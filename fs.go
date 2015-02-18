// +build fs

package readeef

import "fmt"

//go:generate webfw-fs -output fs_files.go -package readeef -format -build-tags fs templates/... static/js/... static/images/... static/css/...

func init() {
	_, err := addFiles()

	if err != nil {
		panic(fmt.Sprintf("Error adding files: %v\n", err))
	}
}
