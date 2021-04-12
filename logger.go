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

var defaultPrintLogger = printfLogger(log.New(os.Stdout, "", defaultLFlag))

func printfLogger(l interface{ Printf(string, ...interface{}) }) Logger {
	return &defaultLogger{
		logger: l,
	}
}

type defaultLogger struct {
	logger interface{ Printf(string, ...interface{}) }
}

func (l *defaultLogger) Error(format string, params ...interface{}) {
	l.logger.Printf(format, params...)
}

func (l *defaultLogger) Info(format string, params ...interface{}) {
	l.logger.Printf(format, params...)
}

func (l *defaultLogger) Debug(format string, params ...interface{}) {
	l.logger.Printf(format, params...)
}
