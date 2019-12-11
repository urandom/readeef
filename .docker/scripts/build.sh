#/bin/sh

set -x
set -e

# go get github.com/golang/dep/cmd/dep
go get github.com/urandom/readeef/cmd/readeef
# go mod tidy
pwd
ls -l

readeef -h

# dep ensure