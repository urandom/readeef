// +build !nofs

package readeef

//go:generate go run ./cmd/readeef-static-locator/main.go -output file.list
//go:generate embed -output fs_files.go -package-name readeef -build-tags !nofs -input file.list
