package proxy

import (
	"fmt"
)

type customLogger struct {
	logFunc LogFunc
	client  *client
}

func (l *customLogger) Errorf(format string, v ...interface{}) {}

func (l *customLogger) Warnf(format string, v ...interface{}) {
	if l.logFunc != nil && l.client != nil {
		l.logFunc("warn", fmt.Sprintf(format, v...), l.client.getCurrentProxy())
	}
}

func (l *customLogger) Debugf(format string, v ...interface{}) {}
func (l *customLogger) Infof(format string, v ...interface{})  {}
