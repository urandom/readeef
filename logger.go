package readeef

import (
	"io"
	"log"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/urandom/readeef/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

type logger struct {
	*logrus.Logger
}

func NewLogger(cfg config.Log) Logger {
	logger := logger{Logger: logrus.New()}

	var writer io.Writer
	if cfg.File == "-" {
		writer = os.Stderr
	} else {
		writer = &lumberjack.Logger{
			Filename:   cfg.File,
			MaxSize:    20,
			MaxBackups: 5,
			MaxAge:     28,
		}
	}

	logger.Out = writer

	switch cfg.Formatter {
	case "text":
		logger.Formatter = &logrus.TextFormatter{}
	case "json":
		logger.Formatter = &logrus.JSONFormatter{}
	}

	switch cfg.Level {
	case "info":
		logger.Level = logrus.InfoLevel
	case "debug":
		logger.Level = logrus.DebugLevel
	case "error":
		logger.Level = logrus.ErrorLevel
	}

	return logger
}

func (l logger) Print(args ...interface{}) {
	l.Logger.Error(args)
}

func (l logger) Printf(format string, args ...interface{}) {
	l.Logger.Errorf(format, args...)
}

func (l logger) Errorln(args ...interface{}) {
	l.Logger.Errorln(args...)
}

// The logger interface provides some common methods for outputting messages.
// It may be used to exchange the default log.Logger error logger with another
// provider.
type Logger interface {
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Fatalln(v ...interface{})

	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})

	Info(v ...interface{})
	Infof(format string, v ...interface{})
	Infoln(v ...interface{})

	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
	Debugln(v ...interface{})
}

type StandardLogger struct {
	*log.Logger
}

func NewStandardLogger(out io.Writer, prefix string, flag int) StandardLogger {
	return StandardLogger{Logger: log.New(out, prefix, flag)}
}

func (st StandardLogger) Info(v ...interface{}) {
	st.Print(v...)
}

func (st StandardLogger) Infof(format string, v ...interface{}) {
	st.Printf(format, v...)
}

func (st StandardLogger) Infoln(v ...interface{}) {
	st.Println(v...)
}

func (st StandardLogger) Debug(v ...interface{}) {
	st.Print(v...)
}

func (st StandardLogger) Debugf(format string, v ...interface{}) {
	st.Printf(format, v...)
}

func (st StandardLogger) Debugln(v ...interface{}) {
	st.Println(v...)
}
