package readeef

import (
	"log"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

type debug interface {
	Printf(string, ...interface{})
	Println(...interface{})
}

type realDebug struct {
	logger *log.Logger
	config Config
}

type blankDebug struct{}

var Debug debug = blankDebug{}

func InitDebug(logger *log.Logger, config Config) {
	Debug = realDebug{logger: logger, config: config}

	Debug.Println("Initializing debug output")
}

func (d realDebug) Printf(format string, v ...interface{}) {
	if d.config.Readeef.Debug {
		_, file, line, ok := runtime.Caller(2)
		timestamp := time.Now().Format(http.TimeFormat)
		if ok {
			d.logger.Println(file + ":" + strconv.Itoa(line) + " (" + timestamp + ")")
		}
		d.logger.Printf(format, v...)
	}
}

func (d realDebug) Println(v ...interface{}) {
	if d.config.Readeef.Debug {
		_, file, line, ok := runtime.Caller(2)
		timestamp := time.Now().Format(http.TimeFormat)
		if ok {
			d.logger.Println(file + ":" + strconv.Itoa(line) + " (" + timestamp + ")")
		}
		d.logger.Println(v...)
	}
}

func (d blankDebug) Printf(format string, v ...interface{}) {}
func (d blankDebug) Println(v ...interface{})               {}
