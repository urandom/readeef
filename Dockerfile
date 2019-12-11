# docker build -t x0rzkov/readeef:alpine3.10-go1.13 .
FROM golang:alpine3.10

RUN apk add --no-cache bash nano make gcc g++ git ca-certificates musl-dev nodejs npm

COPY . /go/src/github.com/urandom/readeef
WORKDIR /go/src/github.com/urandom/readeef

# RUN go get -u github.com/golang/dep/cmd/dep \
# && npm install --unsafe-perm -g node-gyp webpack-dev-server rimraf webpack typescript @angular/cli \

CMD ["/bin/bash"]
