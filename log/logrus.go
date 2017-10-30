package log

import (
	lg "github.com/Sirupsen/logrus"
	"github.com/urandom/readeef/config"
)

type logrus struct {
	*lg.Logger
}

func WithLogrus(cfg config.Log) Log {
	logger := logrus{Logger: lg.New()}

	logger.Out = cfg.Converted.Writer

	switch cfg.Formatter {
	case "text", "":
		logger.Formatter = &lg.TextFormatter{DisableTimestamp: true}
	case "json":
		logger.Formatter = &lg.JSONFormatter{}
	}

	switch cfg.Level {
	case "info":
		logger.Level = lg.InfoLevel
	case "debug":
		logger.Level = lg.DebugLevel
	case "error", "":
		logger.Level = lg.ErrorLevel
	}

	return logger
}

func (l logrus) Print(args ...interface{}) {
	l.Logger.Error(args)
}

func (l logrus) Printf(format string, args ...interface{}) {
	l.Logger.Errorf(format, args...)
}

func (l logrus) Errorln(args ...interface{}) {
	l.Logger.Errorln(args...)
}
