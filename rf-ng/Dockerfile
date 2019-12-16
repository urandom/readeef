# Use a Node.js image and assign it as our build
FROM mhart/alpine-node:12 as build

# Install system dependencies, mostly for libpng
RUN apk --no-cache update \
    && apk --no-cache add g++ make bash zlib-dev libpng-dev nano bash \
    &&  rm -fr /var/cache/apk/*

COPY . /opt/rf-ng

WORKDIR /opt/rf-ng
RUN npm install && \
    npm run-script build

VOLUME /opt/rf-ng/build

CMD ["npm", "start"]