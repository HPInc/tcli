// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package config

import (
	"log"
	"net/http"
)

// level declaration
type Level uint32

/*
30 40 Black
31 41 Red
32 42 Green
33 43 Yellow
34 44 Blue
35 45 Magenta
36 46 Cyan
37 47 White
90 100 Bright Black
91 101 Bright Red
92 102 Bright Green
93 103 Bright Yellow
94 104 Bright Blue
95 105 Bright Magenta
96 106 Bright Cyan
97 107 Bright White
*/
const (
	colorNone    = "\033[0m"
	colorRed     = "\033[0;31m"
	colorGreen   = "\033[0;32m"
	colorYellow  = "\033[0;33m"
	colorBlue    = "\033[0;34m"
	colorMagenta = "\033[0;35m"
	colorCyan    = "\033[0;36m"
	colorWhite   = "\033[0;37m"
)

/*
func main() {
	    fmt.Fprintf(os.Stdout, "Red: \033[0;31m %s None: \033[0m %s", "red string", "colorless string")
	        fmt.Fprintf(os.Stdout, "Red: %s %s None: %s %s", colorRed, "red string", colorNone, "colorless string")
	}
*/

const (
	Info Level = iota
	Verbose
)

type Log struct {
	docProvider DocProvider
	level       Level
	logger      *log.Logger
	plainLogger *log.Logger
}

func InitLogger(l Level) *Log {
	return &Log{
		level:       l,
		logger:      log.Default(),
		plainLogger: log.New(log.Default().Writer(), "", 0),
	}
}

// parse and set doc type
func (l *Log) SetDocType(dt string) {
	l.docProvider = NewDocProvider(dt)
}

func (l *Log) SetLevel(value Level) {
	l.level = value
}

func (l *Log) IsVerbose() bool {
	return l.level >= Verbose
}

func (l *Log) IsInfo() bool {
	return l.level >= Info
}

func (l *Log) Debug(v ...any) {
	if l.IsVerbose() {
		l.logger.Println(v...)
	}
}

func (l *Log) Debugf(format string, v ...any) {
	if l.IsVerbose() {
		l.logger.Printf(format, v...)
	}
}

func (l *Log) Info(v ...any) {
	if l.IsInfo() {
		l.logger.Println(v...)
	}
}

func (l *Log) FatalIf(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func (l *Log) Fatal(v ...any) {
	log.Fatal(v...)
}

func (l *Log) Fatalf(format string, v ...any) {
	log.Fatalf(format, v...)
}

func (l *Log) Errorf(format string, v ...any) {
	l.plainLogger.Printf("%s>>>", colorRed)
	l.logger.Printf(format, v...)
	l.plainLogger.Printf("<<<%s", colorNone)
}

func (l *Log) Successf(format string, v ...any) {
	l.plainLogger.Printf("%s", colorGreen)
	l.logger.Printf(format, v...)
	l.plainLogger.Printf("%s", colorNone)
}

// this function is here for a possible decoration for errors
// right now, its the same as Println
func (l *Log) Error(v ...any) {
	l.plainLogger.Printf("%s>>>", colorRed)
	log.Println(v...)
	l.plainLogger.Printf("<<<%s", colorNone)
}

func (l *Log) Println(v ...any) {
	log.Println(v...)
}

func (l *Log) Printf(format string, v ...any) {
	log.Printf(format, v...)
}

func (l *Log) GetPlainLogger() *log.Logger {
	return l.plainLogger
}

// some specialized logging for http requests and responses
func (l *Log) HttpRequest(req *http.Request) {
	l.docProvider.HttpRequest(req)
}

func (l *Log) HttpResponse(resp *http.Response, body []byte) {
	l.docProvider.HttpResponse(resp, body)
}
