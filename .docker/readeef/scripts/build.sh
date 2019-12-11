#/bin/sh

set -x
set -e

# go get github.com/golang/dep/cmd/dep
# go get github.com/urandom/readeef/cmd/readeef
go mod tidy
pwd
ls -l

# dep ensure
# ls -l

cd rf-ng
npm install --unsafe-perm -g node-gyp webpack-dev-server rimraf webpack typescript @angular/cli
npm install
cd ..

# dep ensure