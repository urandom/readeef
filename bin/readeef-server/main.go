package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"readeef"
	"readeef/api"
	"readeef/web"

	"github.com/urandom/webfw"
)

func main() {
	confpath := flag.String("config", "", "config path")
	host := flag.String("host", "", "server host")
	port := flag.Int("port", 0, "server port")

	flag.Parse()

	cfg, err := readeef.ReadConfig(*confpath)
	if err != nil {
		exitWithError(fmt.Sprintf("Error reading config from path '%s': %v", *confpath, err))
	}

	server := webfw.NewServer(*confpath)
	if *host != "" {
		server.SetHost(*host)
	}

	if *port > 0 {
		server.SetPort(*port)
	}

	dispatcher := server.Dispatcher("/api/")

	logger := log.New(os.Stderr, "", 0)

	if err := api.RegisterControllers(cfg, dispatcher, logger); err != nil {
		exitWithError(err.Error())
	}

	dispatcher = server.Dispatcher("/")
	web.RegisterControllers(dispatcher, "/api/")

	if err := server.ListenAndServe(); err != nil {
		exitWithError(fmt.Sprintf("Error starting server: %s\n", err.Error()))
	}
}

func exitWithError(err string) {
	fmt.Fprintf(os.Stderr, err+"\n")
	os.Exit(1)
}
