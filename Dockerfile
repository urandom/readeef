FROM golang:alpine3.10

MAINTAINER x0rzkov

RUN apk add --no-cache bash nano make gcc g++ git ca-certificates musl-dev nodejs npm sqlite-dev sqlite

COPY . /go/src/github.com/urandom/readeef
WORKDIR /go/src/github.com/urandom/readeef

RUN ./.docker/readeef/scripts/build.sh

# RUN go get -u github.com/golang/dep/cmd/dep \
# && npm install --unsafe-perm -g node-gyp webpack-dev-server rimraf webpack typescript @angular/cli \

# ENTRYPOINT ["./.docker/readeef/scripts/entrypoint.sh"]
# CMD ["dev"]

CMD ["/bin/bash"]