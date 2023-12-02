package logging

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

type Level int

const (
	TRACE = iota
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
	PANIC
)

func NewLevel(l string) Level {
	switch strings.ToLower(l) {
	case "trace":
		return TRACE
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn":
		return WARN
	case "error":
		return ERROR
	case "fatal":
		return FATAL
	case "panic":
		return PANIC
	default:
		return INFO
	}
}

func (l Level) LogLevel() log.Level {
	switch l {
	case TRACE:
		return log.TraceLevel
	case DEBUG:
		return log.DebugLevel
	case INFO:
		return log.InfoLevel
	case WARN:
		return log.WarnLevel
	case ERROR:
		return log.ErrorLevel
	case FATAL:
		return log.FatalLevel
	case PANIC:
		return log.PanicLevel
	default:
		return log.InfoLevel
	}
}

func SetupLogging(file string, level string) *log.Logger {
	l := NewLevel(level)
	f, _ := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	return &log.Logger{
		Out: f, Level: l.LogLevel(),
		Formatter: &log.TextFormatter{DisableColors: true, FullTimestamp: true, PadLevelText: true},
	}
}
