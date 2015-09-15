package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/urandom/readeef"
	"github.com/urandom/readeef/api"
	_ "github.com/urandom/readeef/content/sql/db/postgres"
	"github.com/urandom/readeef/web"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/middleware"
	"gopkg.in/natefinch/lumberjack.v2"
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

	if *serverconfpath == "" {
		server.Config.Session.IgnoreURLPrefix = []string{"/v2/fever", "/v12/tt-rss"}
		server.Config.I18n.Languages = []string{"en", "bg"}
		server.Config.I18n.IgnoreURLPrefix = []string{"/dist", "/js", "/css", "/images", "/proxy"}
	}

	var accessWriter io.Writer
	if cfg.Logger.AccessFile == "-" {
		accessWriter = os.Stdout
	} else {
		accessWriter = &lumberjack.Logger{
			Filename:   cfg.Logger.AccessFile,
			MaxSize:    20,
			MaxBackups: 5,
			MaxAge:     28,
		}
	}

	accessLogger := webfw.NewStandardLogger(accessWriter, "", 0)

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
	os.Exit(0)
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	rand.Seed(time.Now().UnixNano())
}
