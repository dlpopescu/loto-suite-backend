package logging

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

var logger *Logger
var loggerOnce sync.Once

func getLogger() *Logger {
	loggerOnce.Do(func() {
		_, filename, _, ok := runtime.Caller(0)
		if ok {
			loggingDir := filepath.Dir(filename)
			logger = &Logger{dir: filepath.Join(loggingDir, "logs")}
		} else {
			logger = &Logger{dir: filepath.Join("logging", "logs")}
		}

		os.MkdirAll(logger.dir, 0755)
	})

	return logger
}

func DebugBe(message string) {
	getLogger().Debug("BE", message, getCallerInfoEx())
}

func InfoBe(message string) {
	getLogger().Info("BE", message, getCallerInfoEx())
}

func WarnBe(message string) {
	getLogger().Warn("BE", message, getCallerInfoEx())
}

func ErrorBe(message string) {
	getLogger().Error("BE", message, getCallerInfoEx())
}

func FatalBe(message string) {
	getLogger().Fatal("BE", message, getCallerInfoEx(), "")
}

// func DebugFe(callerInfo string, message string) {
// 	getLogger().write("FE", LogTypeDebug, message, callerInfo, "na")
// }

// func InfoFe(callerInfo string, message string) {
// 	getLogger().write("FE", LogTypeInfo, message, callerInfo, "na")
// }

// func WarnFe(callerInfo string, message string) {
// 	getLogger().write("FE", LogTypeWarn, message, callerInfo, "na")
// }

func ErrorFe(message string, callerInfo string) {
	getLogger().Error("FE", message, callerInfo)
}

// func FatalFe(callerInfo string, message string, stackTrace string) {
// 	getLogger().write("FE", LogTypeFatal, message, callerInfo, stackTrace)
// }
