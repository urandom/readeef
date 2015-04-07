readeef
=======

readeef is a self-hosted feed aggregator. Similar to Google Reader, but on your own server.

For a less brief description, [click here](http://www.sugr.org/en/products/readeef).
Some screenshots may also be had [on this page](http://www.sugr.org/en/products/readeef#gallery)

Quick start
===========

readeef is written in Go, and as of September 2014, requires at least version 1.3 of the language. The currently supported databases are PostgreSQL, and SQLite. SQLite support is only built if CGO is enabled. The later is not recommended, as locking problems will occur.

Three binaries may be built from the sources. The first is a user administration script, which can be used to add, remove and modify users. It is not necessary have this binary, as readeef will create an 'admin' user with password 'admin', if such a user doesn't already exist:

> go build github.com/urandom/readeef/cmd/readeef-user-admin

The second binary is the standalone server. Unless readeef is being added to an existing golang server setup, it should be built as well. Since readeef uses bleve for FTS capabilities, bleve-specific tags (e.g.: leveldb, cld2, etc) should be passed here.

> go build github.com/urandom/readeef/cmd/readeef-server

Unless you are using SQLite, readeef will need to be configured as well. A minimal configuration file might be something like this:

```
[db]
    driver = postgres
    connect = host=/var/run/postgresql user=postgresuser dbname=readeefdbname
[timeout]
    connect = 1s
    read-write = 2s
[auth]
    secret = someverylongsecretstring
```

You are now ready to add a user to the system. Turn the user into an administrator to be able to add more users via the web interface. You may skip this step if the default admin user is suitable for you.

> ./readeef-user-admin -config $CONFIG_FILE add $USER_LOGIN $USER_PASS

> ./readeef-user-admin -config $CONFIG_FILE set $USER_LOGIN admin true

The standalone server may take two config files. The first is the readeef configuration file, and the other is the server configuration. The later one is optional. The default server configuration may be seen in the source code of this file: [webfw/config.go](https://github.com/urandom/webfw/blob/master/config.go#L120). The server will need to be started in the same directory that contains the 'static' and 'templates' directories, typically the checkout itself.

> ./readeef-server -server-config $SERVER_CONFIG_FILE -readeef-config $CONFIG_FILE

Unless the server has been built with the 'nofs' tag, the client-side libraries will need to be fetched. This is best done with bower. Make sure the _.bowerrc_ file, provided with the sources, is in the same directory that contains the 'static' directory. In there, just run the following:

> bower update


"But I just want to try it"
===========================

    # Install the server in $GOPATH/.bin/
    go get github.com/urandom/readeef/cmd/readeef-server
    # Run it using the default settings
    readeef-server
    
The server will run on port 8080, and you may login using the user 'admin' and password 'admin', using SQLite (if CGO is enabled)
