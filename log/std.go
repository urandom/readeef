package log

import (
	"log"

	"github.com/urandom/readeef/config"
)

type stdLogger struct {
	*log.Logger
}

// WithStd creates a logger that uses the stdlib log facilities.
func WithStd(cfg config.Log) Log {
	return stdLogger{Logger: log.New(cfg.Converted.Writer, cfg.Converted.Prefix, 0)}
}

func (st stdLogger) Info(v ...interface{}) {
	st.Print(v...)
}

func (st stdLogger) Infof(format string, v ...interface{}) {
	st.Printf(format, v...)
}

func (st stdLogger) Infoln(v ...interface{}) {
	st.Println(v...)
}

func (st stdLogger) Debug(v ...interface{}) {
	st.Print(v...)
}

func (st stdLogger) Debugf(format string, v ...interface{}) {
	st.Printf(format, v...)
}

func (st stdLogger) Debugln(v ...interface{}) {
	st.Println(v...)
}
