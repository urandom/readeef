#/bin/sh

set -x
set -e

# build binaries
go mod tidy
go get github.com/urandom/embed/cmd/embed
# go get github.com/urandom/readeef/cmd/readeef
go get github.com/urandom/readeef/cmd/readeef-static-locator

# requirements for building ui
cd rf-ng
npm install --unsafe-perm -g node-gyp webpack-dev-server rimraf webpack typescript @angular/cli
npm install
cd ..

make all