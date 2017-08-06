package log

import (
	"io"
	"log"
)

type stdLogger struct {
	*log.Logger
}

// WithStd creates a logger that uses the stdlib log facilities.
func WithStd(out io.Writer, prefix string, flag int) Log {
	return stdLogger{Logger: log.New(out, prefix, flag)}
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
