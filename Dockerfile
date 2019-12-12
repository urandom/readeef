FROM golang:alpine3.10 AS builder
MAINTAINER x0rzkov

RUN apk add --no-cache bash nano make gcc g++ git ca-certificates musl-dev nodejs npm sqlite-dev sqlite

COPY . /go/src/github.com/urandom/readeef
WORKDIR /go/src/github.com/urandom/readeef

RUN ./.docker/readeef/scripts/build.sh

FROM alpine:3.10 AS runtime
MAINTAINER x0rzkov

ARG TINI_VERSION=${TINI_VERSION:-"v0.18.0"}

# Install tini to /usr/local/sbin
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini-muslc-amd64 /usr/local/sbin/tini

# Install runtime dependencies & create runtime user
RUN apk --no-cache --no-progress add ca-certificates git libssh2 openssl sqlite \
 && chmod +x /usr/local/sbin/tini && mkdir -p /opt \
 && adduser -D readeef -h /opt/readeef -s /bin/sh \
 && su readeef -c 'cd /opt/readeef; mkdir -p bin config data'

# Copy readeef binary to /opt/readeef/bin
COPY --from=builder /go/src/github.com/urandom/readeef/readeef /opt/readeef/bin/readeef
COPY .docker/readeef/config/readeef.toml /opt/readeef/config/readeef.toml
ENV PATH $PATH:/opt/readeef/bin

# Container configuration
EXPOSE 8080
VOLUME ["/opt/readeef/data"]
ENTRYPOINT ["tini", "-g", "--"]
CMD ["/opt/readeef/bin/readeef", "-config=/opt/readeef/config/readeef.toml"]
# ENTRYPOINT ["./.docker/readeef/scripts/entrypoint.sh"]
# CMD ["dev"]
