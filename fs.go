// +build !nofs

package readeef

import "fmt"

//go:generate go run ./cmd/readeef-static-locator/main.go -output file.list
//go:generate webfw-fs -output fs_files.go -package readeef -format -build-tags !nofs -input file.list

func init() {
	_, err := addFiles()

	if err != nil {
		panic(fmt.Sprintf("Error adding files: %v\n", err))
	}
}
