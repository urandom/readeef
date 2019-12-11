# Use a Node.js image and assign it as our build
FROM mhart/alpine-node:12 as build

# Install system dependencies, mostly for libpng
RUN apk --no-cache update \ 
  && apk --no-cache add g++ make bash zlib-dev libpng-dev \
  &&  rm -fr /var/cache/apk/*

