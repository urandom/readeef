readeef
=======

readeef is a self-hosted feed aggregator. Similar to Google Reader, but on your own server.

For a more detailed description, [click here](http://www.sugr.org/en/products/readeef).
Some screenshots may also be had [on this page](http://www.sugr.org/en/products/readeef#gallery)

Quick start
===========

readeef is written in Go, and as of September 2014, requires at least version 1.3 of the language. The currently supported databases are PostgreSQL, and SQLite. SQLite support is only built if CGO is enabled. The later is not recommended, as locking problems will occur.

Three binaries may be built from the sources. The first binary is the standalone server. Unless readeef is being added to an existing golang server setup, it should be built as well. Since readeef can use bleve for FTS capabilities, bleve-specific tags (e.g.: libstemmer, cld2, etc) should be passed here.

> go build github.com/urandom/readeef/cmd/readeef-server

Unless you are using SQLite, readeef will need to be configured as well. A minimal configuration file might be something like this:

```
[db]
    driver = postgres
    connect = host=/var/run/postgresql user=postgresuser dbname=readeefdbname
```

You may provide The standalone server with a config files. The default server configuration is documented in godoc.org under the variable: [DefaultCfg](http://godoc.org/github.com/urandom/readeef#pkg-variables). The server will need to be started in the same directory that contains the 'static' and 'templates' directories, typically the checkout itself.

> ./readeef-server -config $CONFIG_FILE

If the server has been built with the 'nofs' tag, the client-side libraries will need to be fetched. This is best done with bower. Make sure the _.bowerrc_ file, provided with the sources, is in the same directory that contains the 'static' directory. In there, just run the following:

> bower update

The second is a user administration script, which can be used to add, remove and modify users. It is not necessary to have this binary, as readeef will create an 'admin' user with password 'admin', if such a user doesn't already exist:

> go build github.com/urandom/readeef/cmd/readeef-user-admin

You can now use the script to add, remove and edit users

> \# Adding a user

> ./readeef-user-admin -config $CONFIG_FILE add $USER_LOGIN $USER_PASS

> \# Turning a user into an admin

> ./readeef-user-admin -config $CONFIG_FILE set $USER_LOGIN admin true

The third is a search index management script, which can be used to re-index all articles in the database. It is usually not necessary have this binary, as articles are indexed when they are added to the database. It might be useful if you switch from one search provider to another:

> go build github.com/urandom/readeef/cmd/readeef-search-index

> ./readeef-search-index -config $CONFIG_FILE


"But I just want to try it"
===========================

    # Install the server in $GOPATH/.bin/
    go get github.com/urandom/readeef/cmd/readeef-server
    # Run it using the default settings
    readeef-server
    
The server will run on port 8080, and you may login using the user 'admin' and password 'admin', using SQLite (if CGO is enabled)
