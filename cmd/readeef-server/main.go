package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/api"
	"github.com/urandom/readeef/web"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/middleware"
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

	logger := readeef.NewLogger(cfg)
	defer func() {
		if rec := recover(); rec != nil {
			stack := debug.Stack()
			logger.Fatalf("Fatal error: %v\n%s\n", rec, stack)
		}
	}()

	server := webfw.NewServer(*serverconfpath)
	if *address != "" {
		server.Address = *address
	}

	if *port > 0 {
		server.Port = *port
	}

	accessLogger := webfw.NewStandardLogger(&lumberjack.Logger{
		Dir:        ".",
		NameFormat: cfg.Logger.AccessFile,
		MaxSize:    10000000,
		MaxBackups: 5,
		MaxAge:     28,
	}, "", 0)

	dispatcher := server.Dispatcher("/api/")
	dispatcher.Logger = logger
	dispatcher.RegisterMiddleware(middleware.Logger{AccessLogger: accessLogger})

	if err := api.RegisterControllers(cfg, dispatcher, logger); err != nil {
		exitWithError(err.Error())
	}

	dispatcher = server.Dispatcher("/")
	dispatcher.Logger = logger
	dispatcher.RegisterMiddleware(middleware.Logger{AccessLogger: accessLogger})

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
