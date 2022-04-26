package pep

//include explanation on how to output logs to their file of

import (
	"fmt"
	"io"
	"log"
)

// Logger is an interface describing the logging contract that must be met in order for log events to propagate from the OnePEP code
type Logger interface {
	Info(args ...interface{})
	Debug(args ...interface{})
	Error(args ...interface{})
	//Trace(args ...interface{})
}

// Level denotes the numeric representation of the logging level
type Level int

// Returns a string representation of the logging `Level`
func (l Level) String() string {
	return [...]string{"LevelDebug", "LevelInfo", "LevelError"}[l]
}

// Logging level definitions
const (
	LevelNotSet Level = iota
	LevelDebug
	LevelInfo
	LevelError
)

// OnePEPLogger is the default logger for OnePEP logs
type OnePEPLogger struct {
	LogLevel Level
	log      *log.Logger
}

// NewOnePEPLogger creates a OnePEPLogger using the provided level and output `io.Writer` and returns it to the caller
func NewOnePEPLogger(level Level, output io.Writer) (*OnePEPLogger, error) {
	var logger *OnePEPLogger

	if level > LevelError || level < LevelDebug {
		return nil, fmt.Errorf("invalid logging level %d provided, use Logging level definitions %s, %s, or %s", level, LevelDebug.String(), LevelInfo.String(), LevelError.String())
	}

	l := log.New(output, "INFO: ", log.Lshortfile)
	logger = &OnePEPLogger{
		LogLevel: level,
		log:      l,
	}

	return logger, nil
}

// Debug prints debug level logs
func (logger *OnePEPLogger) Debug(v ...interface{}) {
	if logger.LogLevel <= LevelDebug {
		logger.logLn(LevelDebug, v...)
	}
}

// Info prints info level logs
func (logger *OnePEPLogger) Info(v ...interface{}) {
	if logger.LogLevel <= LevelInfo {
		logger.logLn(LevelInfo, v...)
	}
}

// Error prints error level logs
func (logger *OnePEPLogger) Error(v ...interface{}) {
	if logger.LogLevel <= LevelError {
		logger.logLn(LevelError, v...)
	}
}

func (logger *OnePEPLogger) logLn(level Level, v ...interface{}) {
	switch level {
	case LevelDebug:
		if !(logger.log.Prefix() == "DEBUG: ") {
			logger.log.SetPrefix("DEBUG: ")
		}
	case LevelInfo:
		if !(logger.log.Prefix() == "INFO: ") {
			logger.log.SetPrefix("INFO: ")
		}
	case LevelError:
		if !(logger.log.Prefix() == "ERROR: ") {
			logger.log.SetPrefix("ERROR: ")
		}
	}
	// using Output(calldepth int, s string) to specify call depth
	_ = logger.log.Output(3, fmt.Sprint(v...))
}
