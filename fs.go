// +build !nofs

package readeef

//go:generate embed -output fs_files.go -package-name readeef -build-tags !nofs -fallback rf-ng/ui/... templates/raw.tmpl templates/goose-format-result.tmpl
