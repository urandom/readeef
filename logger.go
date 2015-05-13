package readeef

import (
	"io"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/urandom/webfw"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	*logrus.Logger
}

func NewLogger(cfg Config) webfw.Logger {
	logger := Logger{Logger: logrus.New()}

	var writer io.Writer
	if cfg.Logger.File == "-" {
		writer = os.Stderr
	} else {
		writer = &lumberjack.Logger{
			Filename:   cfg.Logger.File,
			MaxSize:    20,
			MaxBackups: 5,
			MaxAge:     28,
		}
	}

	logger.Out = writer

	switch cfg.Logger.Formatter {
	case "text":
		logger.Formatter = &logrus.TextFormatter{}
	case "json":
		logger.Formatter = &logrus.JSONFormatter{}
	}

	switch cfg.Logger.Level {
	case "info":
		logger.Level = logrus.InfoLevel
	case "debug":
		logger.Level = logrus.DebugLevel
	case "error":
		logger.Level = logrus.ErrorLevel
	}

	return logger
}

func (l Logger) Print(args ...interface{}) {
	l.Logger.Error(args)
}

func (l Logger) Printf(format string, args ...interface{}) {
	l.Logger.Errorf(format, args...)
}

func (l Logger) Errorln(args ...interface{}) {
	l.Logger.Errorln(args...)
}
