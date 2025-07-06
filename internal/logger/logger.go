package logger

import (
	"io"
	"log"
	"os"
)

var (
	// Debug mode status
	debugMode bool
	// Standard logger
	stdLog *log.Logger
)

// Init initializes the logger
func Init(debug bool) {
	debugMode = debug

	// Configure output destination
	var output io.Writer
	if debugMode {
		output = os.Stdout
	} else {
		output = io.Discard // No output when not in debug mode
	}

	stdLog = log.New(output, "", log.LstdFlags)
}

// Println outputs log messages only in debug mode
func Println(v ...interface{}) {
	if stdLog == nil {
		Init(false)
	}
	stdLog.Println(v...)
}

// Printf outputs formatted log messages only in debug mode
func Printf(format string, v ...interface{}) {
	if stdLog == nil {
		Init(false)
	}
	stdLog.Printf(format, v...)
}

// IsDebugMode returns the current debug mode status
func IsDebugMode() bool {
	return debugMode
}
