package logs

import (
	"os"
)

var DefaultLogger *Logger

func SetDefaultLogger(logger *Logger) {
	DefaultLogger = logger
	if logger.Name == "" {
		logger.Name = "默认"
	}
}

func init() {
	DefaultLogger = NewLoggerWithName("默认", DebugLevel, os.Stdout)
}

func Debug(content ...interface{}) {
	DefaultLogger.Debug(content...)
}
func Info(content ...interface{}) {
	DefaultLogger.Info(content...)
}
func Warn(content ...interface{}) {
	DefaultLogger.Warn(content...)
}
func Error(content ...interface{}) {
	DefaultLogger.Error(content...)
}
func Fatal(content ...interface{}) {
	DefaultLogger.Fatal(content...)
}
