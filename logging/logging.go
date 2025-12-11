package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
)

var logPackageName string
var logPackageOnce sync.Once

// ts=<timestamp>^s=<source>^lt=<log_type>^m=<message>^ci=<caller_info>^st=<stack_trace>
const logLineFormat = "ts=%s^" +
	"s=%s^" +
	"lt=%s^" +
	"m=%s^" +
	"ci=%s^" +
	"st=%s" +
	"\n"

var logger *Logger
var loggerOnce sync.Once

func getLogger() *Logger {
	loggerOnce.Do(func() {
		logger = &Logger{dir: filepath.Join(".", "logs")}
		os.MkdirAll(logger.dir, 0755)
	})

	return logger
}

func Debug(source string, message string) {
	getLogger().debug(source, message)
}

func Info(source string, message string) {
	getLogger().info(source, message)
}

func Warn(source string, message string) {
	getLogger().warn(source, message)
}

func Error(source string, err error, callerInfo string) {
	if strings.TrimSpace(callerInfo) == "" {
		callerInfo = handleEmptyValue(getCallerInfoEx())
	}

	getLogger().error(source, err.Error(), callerInfo)
}

func getCallerInfoEx() string {
	const maxCallerInfoSkip = 10

	logPackageOnce.Do(func() {
		funcPC := reflect.ValueOf(getCallerInfoEx).Pointer()
		if fn := runtime.FuncForPC(funcPC); fn != nil {
			fullName := fn.Name()
			if lastDot := strings.LastIndex(fullName, "."); lastDot != -1 {
				logPackageName = fullName[:lastDot] + "."
			}
		}
	})

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
		if logPackageName != "" && strings.HasPrefix(funcName, logPackageName) {
			continue
		}

		return fmt.Sprintf("%s:%d:%s", filepath.Base(file), line, filepath.Base(funcName))
	}

	return ""
}

func GetLogDir() string {
	return getLogger().dir
}
