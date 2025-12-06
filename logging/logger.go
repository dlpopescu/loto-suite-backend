package logging

import (
	"fmt"
	"loto-suite/backend/generics"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Logger struct {
	dir   string
	mutex sync.Mutex
}

func (l *Logger) Debug(source string, message string, callerInfo string) {
	l.Write(source, LogTypeDebug, message, callerInfo, "na")
}

func (l *Logger) Info(source string, message string, callerInfo string) {
	l.Write(source, LogTypeInfo, message, callerInfo, "na")
}

func (l *Logger) Warn(source string, message string, callerInfo string) {
	l.Write(source, LogTypeWarn, message, callerInfo, "na")
}

func (l *Logger) Error(source string, message string, callerInfo string) {
	l.Write(source, LogTypeError, message, callerInfo, "na")
}

func (l *Logger) Fatal(source string, message string, callerInfo string, stackTrace string) {
	if source == "BE" && strings.TrimSpace(stackTrace) == "" {
		stackTrace = l.getStackTrace()
	}

	l.Write(source, LogTypeFatal, message, callerInfo, stackTrace)
}

func (l *Logger) Write(source string, logType LogType, message string, callerInfo string, stackTrace string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	date := time.Now().Format(generics.GoDateFormat)
	logFileName := fmt.Sprintf("app_%s.log", date)
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

func (l *Logger) getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

func handleEmptyValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "na"
	}

	return value
}
