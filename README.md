readeef
=======

readeef is a self-hosted feed aggregator. Similar to Google Reader, but on your own server.

For a more detailed description, [click here](http://www.sugr.org/en/products/readeef).
Some screenshots may also be had [on this page](http://www.sugr.org/en/products/readeef#gallery)

Quick start
===========

readeef is written in Go, and as of October 2017, requires at least version 1.8 of the language. The currently supported databases are PostgreSQL, and SQLite. SQLite support is only built if CGO is enabled. The later is not recommended, as locking problems will occur.

A single binary may be built from the sources. It current contains three subcommands, one for starting the server, one for rebuinding the search index (while the server is stopped), and an administrative command, for manipulating users. Since readeef can use bleve for FTS capabilities, bleve-specific tags (e.g.: libstemmer, cld2, etc) should be passed here.

> go build github.com/urandom/readeef/cmd/readeef

Unless you are using SQLite, readeef will need to be configured as well. readeef uses TOML for configuration. A minimal configuration file might be something like this:

```
[db]
    driver = "postgres"
    connect = "host=/var/run/postgresql user=postgresuser dbname=readeefdbname"
```

You may provide the standalone server with a config files. The default server configuration is documented in godoc.org under the variable: [DefaultCfg](http://godoc.org/github.com/urandom/readeef/config#pkg-variables).

> ./readeef -config $CONFIG_FILE server

The source comes with an embedded UI using angular 4. A different UI may be provided by providing a path to it via the following configuration directive:

> [ui]
>      path = "/path/to/a/different/ui"

All three subcommands come with a comprehensive usage text:

> readeef search-index --help

### Adding a user

As a first step, you might want to add a new user to the system, using the 'user-admin' subcommand:

> readeef -config $CONFIG_PATH user-admin add $USER_LOGIN $USER_PASS

### Turning a user into an admin

You might then want to turn that user into an administrator:

> ./readeef -config $CONFIG_FILE user-admin set $USER_LOGIN admin true

"But I just want to try it"
===========================

    # Install the server in $GOPATH/.bin/
    go get github.com/urandom/readeef/cmd/readeef
    # Run it using the default settings
    readeef server
    
The server will run on port 8080, and you may login using the user 'admin' and password 'admin', using SQLite (if CGO is enabled)
