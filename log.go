package main

import (
	"io/ioutil"
	"log"
	"os"
)

type localLog struct {
	infoLog    *log.Logger
	warningLog *log.Logger
	errorLog   *log.Logger
	fatalLog   *log.Logger
}

const defaultFlags = log.Ltime

const (
	levelInfo  = "info"
	levelWarn  = "warn"
	levelError = "error"
)

func newLogger(level string) *localLog {
	//fatal will always output to stderr no matter the logLevel
	infoWriter := ioutil.Discard
	warnWriter := ioutil.Discard
	errorWriter := ioutil.Discard
	fatalWriter := os.Stderr

	switch level {
	case levelInfo:
		infoWriter = os.Stdout
		fallthrough
	case levelWarn:
		warnWriter = os.Stdout
		fallthrough
	case levelError:
		errorWriter = os.Stderr
	}

	Logger := localLog{
		infoLog:    log.New(infoWriter, "[INFO]\t", defaultFlags),
		warningLog: log.New(warnWriter, "[WARN]\t", defaultFlags),
		errorLog:   log.New(errorWriter, "[ERROR]\t", defaultFlags),
		fatalLog:   log.New(fatalWriter, "[FATAL]\t", defaultFlags),
	}
	return &Logger
}

func (l *localLog) Info(message ...string) {
	l.infoLog.Print(message)
}

func (l *localLog) Warning(message ...string) {
	l.warningLog.Print(message)
}

func (l *localLog) Error(err error) {
	l.errorLog.Print(err.Error())
}

func (l *localLog) Fatal(message ...string) {
	l.fatalLog.Fatal(message)
}
