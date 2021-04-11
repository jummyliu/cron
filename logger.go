package cron

import (
	"log"
	"os"
)

type Logger interface {
	Error(format string, params ...interface{})
	Info(format string, params ...interface{})
	Debug(format string, params ...interface{})
}

const (
	defaultLFlag = log.Ldate | log.Ltime | log.Lmicroseconds
	allLFlag     = log.Ldate | log.Ltime | log.Lmicroseconds | log.Llongfile | log.Lshortfile
)

var defaultLogger = printfLogger(log.New(os.Stdout, "", defaultLFlag))

func printfLogger(l interface{ Printf(string, ...interface{}) }) Logger {
	return &DefaultLogger{
		logger: l,
	}
}

type DefaultLogger struct {
	logger interface{ Printf(string, ...interface{}) }
}

func (l *DefaultLogger) Error(format string, params ...interface{}) {
	l.logger.Printf(format, params...)
}

func (l *DefaultLogger) Info(format string, params ...interface{}) {
	l.logger.Printf(format, params...)
}

func (l *DefaultLogger) Debug(format string, params ...interface{}) {
	l.logger.Printf(format, params...)
}
