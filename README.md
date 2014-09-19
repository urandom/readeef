readeef
=======

readeef is a self-hosted feed aggregator. Similar to Google Reader, but on your own server.

For a less brief description, [click here](http://www.sugr.org/en/products/readeef).
Some screenshots may also be had [on this page](http://www.sugr.org/en/products/readeef#gallery)

Quick start
===========

readeef is written in Go, and as of September 2014, requires at least version 1.3 of the language. It also uses build tags to specify which database support to be built it. The currently supported databases are PostgreSQL ('postgres' tag), and SQLite ('sqlite3' tag). The later is not recommended, as locking problems will occur. 
Two binaries may be built from the sources. The first is a user administration script, which has to be used to add the first user to the system. It may be built using the following command:

> go build -tags postgres github.com/urandom/readeef/bin/readeef-user-admin

The second binary is the standalone server. Unless readeef is being added to an existing golang server setup, it should be built as well.

> go build -tags postgres github.com/urandom/readeef/bin/readeef-server

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

You are now ready to add the first user to the system. Turn the user into an administrator to be able to add more users via the web interface.

> ./readeef-user-admin -config $CONFIG_FILE add $USER_LOGIN $USER_PASS
> ./readeef-user-admin -config $CONFIG_FILE set $USER_LOGIN admin true

The standalone server may take two config files. The first is the readeef configuration file, and the other is the server configuration. The later one is optional. The default server configuration may be seen in the source code of this file: [webfw/config.go](https://github.com/urandom/webfw/blob/master/config.go#L120). The server will need to be started in the same directory that contains the 'static' and 'templates' directories, typically the checkout itself.

> ./readeef-server -server-config $SERVER_CONFIG_FILE -readeef-config $CONFIG_FILE 2> error.log > access.log

In order for the web interface to actually work, the client-side libraries will need to be fetched. This is best done with bower. Make sure the _.bowerrc_ file, provided with the sources, is in the same directory that contains the 'static' directory. In there, just run the following:

> bower update
