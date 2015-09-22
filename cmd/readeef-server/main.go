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
	configpath := flag.String("config", "", "readeef config path")

	flag.Parse()

	cfg, err := readeef.ReadConfig(*configpath)
	if err != nil {
		exitWithError(fmt.Sprintf("Error reading config from path '%s': %v", *configpath, err))
	}

	if len(cfg.Config.Session.IgnoreURLPrefix) == 0 {
		cfg.Config.Session.IgnoreURLPrefix = []string{"/v2/fever", "/v12/tt-rss"}
	}
	if len(cfg.Config.I18n.Languages) == 0 {
		cfg.Config.I18n.Languages = []string{"en", "bg"}
	}
	if len(cfg.Config.I18n.IgnoreURLPrefix) == 0 {
		cfg.Config.I18n.IgnoreURLPrefix = []string{"/dist", "/js", "/css", "/images", "/proxy"}
	}

	logger := readeef.NewLogger(cfg)
	defer func() {
		if rec := recover(); rec != nil {
			stack := debug.Stack()
			logger.Fatalf("Fatal error: %v\n%s\n", rec, stack)
		}
	}()

	server := webfw.NewServerWithConfig(cfg.Config)

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
