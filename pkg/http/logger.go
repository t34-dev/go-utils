package http

import (
	"fmt"
	"strings"
)

// restyLogger implements the Logger interface for Resty
type restyLogger struct {
	logFunc LogFunc
}

func (l *restyLogger) Errorf(format string, v ...interface{}) {
	l.logFunc("error", fmt.Sprintf(format, v...))
}

func (l *restyLogger) Warnf(format string, v ...interface{}) {
	l.logFunc("warn", fmt.Sprintf(format, v...))
}

func (l *restyLogger) Debugf(format string, v ...interface{}) {
	l.logFunc("debug", fmt.Sprintf(format, v...))
}

type customLogger struct {
	logFunc    LogFunc
	logRetries bool
}

func (l *customLogger) Errorf(format string, v ...interface{}) {
	if l.logRetries && strings.Contains(format, "Retrying request") {
		l.logFunc("error", fmt.Sprintf(format, v...))
	}
}

func (l *customLogger) Warnf(format string, v ...interface{}) {
	if l.logRetries && strings.Contains(format, "Retrying request") {
		l.logFunc("warn", fmt.Sprintf(format, v...))
	}
}

func (l *customLogger) Debugf(format string, v ...interface{}) {
	// Можно добавить дополнительную логику фильтрации, если нужно
}
