package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/natefinch/lumberjack"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/api"
	"github.com/urandom/readeef/web"
	"github.com/urandom/webfw"
)

func main() {
	serverconfpath := flag.String("server-config", "", "server config path")
	readeefconfpath := flag.String("readeef-config", "", "readeef config path")
	address := flag.String("address", "", "local server network address")
	port := flag.Int("port", 0, "server port")

	flag.Parse()

	cfg, err := readeef.ReadConfig(*readeefconfpath)
	if err != nil {
		exitWithError(fmt.Sprintf("Error reading config from path '%s': %v", *readeefconfpath, err))
	}

	logger := logrus.New()
	logger.Out = &lumberjack.Logger{
		Dir:        ".",
		NameFormat: "error-2006-01-02.000.log",
		MaxSize:    10000000,
		MaxBackups: 5,
		MaxAge:     28,
	}

	switch cfg.Logger.Level {
	case "info":
		logger.Level = logrus.InfoLevel
	case "debug":
		logger.Level = logrus.DebugLevel
	}

	server := webfw.NewServer(*serverconfpath)
	if *address != "" {
		server.Address = *address
	}

	if *port > 0 {
		server.Port = *port
	}

	dispatcher := server.Dispatcher("/api/")
	dispatcher.Logger = logger

	if err := api.RegisterControllers(cfg, dispatcher, logger); err != nil {
		exitWithError(err.Error())
	}

	dispatcher = server.Dispatcher("/")
	dispatcher.Logger = logger
	web.RegisterControllers(cfg, dispatcher, "/api/")

	if err := server.ListenAndServe(); err != nil {
		exitWithError(fmt.Sprintf("Error starting server: %s\n", err.Error()))
	}
}

func exitWithError(err string) {
	fmt.Fprintf(os.Stderr, err+"\n")
	os.Exit(1)
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	rand.Seed(time.Now().UnixNano())
}
