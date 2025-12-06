package logging

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
)

const maxCallerInfoSkip = 10

var packageName string
var loggingPackageOnce sync.Once

func getLoggingPackageName() {
	funcPC := reflect.ValueOf(getCallerInfoEx).Pointer()
	if fn := runtime.FuncForPC(funcPC); fn != nil {
		fullName := fn.Name()
		if lastDot := strings.LastIndex(fullName, "."); lastDot != -1 {
			packageName = fullName[:lastDot] + "."
		}
	}
}

func getCallerInfoEx() string {
	loggingPackageOnce.Do(getLoggingPackageName)

	for skip := 1; skip < maxCallerInfoSkip; skip++ {
		pc, file, line, ok := runtime.Caller(skip)
		if !ok {
			break
		}

		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		funcName := fn.Name()
		if packageName != "" && strings.HasPrefix(funcName, packageName) {
			continue
		}

		return fmt.Sprintf("%s:%d:%s", filepath.Base(file), line, filepath.Base(funcName))
	}

	return "na"
}

type LogType string

const (
	LogTypeDebug LogType = "DEBUG"
	LogTypeInfo  LogType = "INFO"
	LogTypeWarn  LogType = "WARN"
	LogTypeError LogType = "ERROR"
	LogTypeFatal LogType = "FATAL"
	LogTypeTrace LogType = "TRACE"
)

// ts=<timestamp>^s=<source>^lt=<log_type>^m=<message>^ci=<caller_info>^st=<stack_trace>
//
// log_type: DEBUG, INFO, WARN, ERROR, FATAL, TRACE
const logLineFormat = "ts=%s^" +
	"s=%s^" +
	"lt=%s^" +
	"m=%s^" +
	"ci=%s^" +
	"st=%s" +
	"\n"
