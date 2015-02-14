package readeef

import (
	"github.com/Sirupsen/logrus"
	"github.com/natefinch/lumberjack"
	"github.com/urandom/webfw"
)

type Logger struct {
	*logrus.Logger
}

func NewLogger(cfg Config) webfw.Logger {
	logger := Logger{Logger: logrus.New()}
	logger.Out = &lumberjack.Logger{
		Dir:        ".",
		NameFormat: cfg.Logger.File,
		MaxSize:    10000000,
		MaxBackups: 5,
		MaxAge:     28,
	}

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
	l.Logger.Errorf(format, args)
}

func (l Logger) Errorln(args ...interface{}) {
	l.Logger.Errorln(args)
}
