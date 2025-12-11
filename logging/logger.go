package logging

import (
	"fmt"
	"loto-suite/backend/generics"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type LogType string

const (
	LogTypeDebug LogType = "DEBUG"
	LogTypeInfo  LogType = "INFO"
	LogTypeWarn  LogType = "WARN"
	LogTypeError LogType = "ERROR"
	LogTypeFatal LogType = "FATAL"
	LogTypeTrace LogType = "TRACE"
)

type Logger struct {
	dir   string
	mutex sync.Mutex
}

func (l *Logger) debug(source string, message string) {
	l.write(source, LogTypeDebug, message, "", "")
}

func (l *Logger) info(source string, message string) {
	l.write(source, LogTypeInfo, message, "", "")
}

func (l *Logger) warn(source string, message string) {
	l.write(source, LogTypeWarn, message, "", "")
}

func (l *Logger) error(source string, message string, callerInfo string) {
	l.write(source, LogTypeError, message, callerInfo, "")
}

func (l *Logger) write(source string, logType LogType, message string, callerInfo string, stackTrace string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	date := time.Now().Format(generics.GoDateFormat)
	logFileName := fmt.Sprintf("be_%s.log", date)
	logFilePath := filepath.Join(l.dir, logFileName)

	timestamp := time.Now().Format(fmt.Sprintf("%s %s", generics.GoDateFormat, generics.GoTimeFormat))

	source = handleEmptyValue(source)
	message = handleEmptyValue(message)
	callerInfo = handleEmptyValue(callerInfo)
	stackTrace = handleEmptyValue(stackTrace)

	logLine := fmt.Sprintf(logLineFormat, timestamp, source, logType, message, callerInfo, stackTrace)
	fmt.Print(logLine)

	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file: %v\n", err)
		return
	}

	defer file.Close()

	if _, err := file.WriteString(logLine); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write to log file: %v\n", err)
	}
}

func handleEmptyValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "NA"
	}

	return value
}
