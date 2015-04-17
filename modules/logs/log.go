package logs

import (
	"fmt"
	"io"
	"os"
	"time"
)

const (
	DebugOn = 0x1
	InfoOn  = 0x3
	WarnOn  = 0x7
	ErrorOn = 0xF
	FatalOn = 0x1F
)

const (
	ErrorLevel = ErrorOn | FatalOn
	WarnLevel  = ErrorLevel | WarnOn
	InfoLevel  = WarnLevel | InfoOn
	DebugLevel = InfoLevel | DebugOn
)

type Logger struct {
	Writers []io.Writer

	Level uint8

	LogWithTime bool

	LastLogTime time.Time
}

func NewLogger(level uint8, w ...io.Writer) *Logger {
	return &Logger{
		Writers:     w,
		Level:       level,
		LogWithTime: true,
		LastLogTime: time.Now(),
	}
}

func (logger *Logger) Log(prefix string, content ...interface{}) {
	now := time.Now()

	if logger.LastLogTime.Day() != now.Day() {
		logger.LastLogTime = now
		for _, writer := range logger.Writers {
			io.WriteString(writer, "#日期# "+fmt.Sprintln(logger.LastLogTime.Format("L2006-01-02")))
		}
	}
	if logger.LogWithTime {
		if logger.LastLogTime.Unix() != now.Unix() {
			logger.LastLogTime = now
			for _, writer := range logger.Writers {
				io.WriteString(writer, "#时间# "+fmt.Sprintln(logger.LastLogTime.Format("L15:04:05")))
			}
		}
	}

	for _, writer := range logger.Writers {
		io.WriteString(writer, "#"+prefix+"# "+fmt.Sprintln(content...))
	}
}

func (logger *Logger) Debug(content ...interface{}) {
	if logger.Level&DebugOn != 0 {
		logger.Log("调试", content...)
	}
}
func (logger *Logger) Info(content ...interface{}) {
	if logger.Level&InfoOn != 0 {
		logger.Log("信息", content...)
	}
}
func (logger *Logger) Warn(content ...interface{}) {
	if logger.Level&WarnOn != 0 {
		logger.Log("警告", content...)
	}
}
func (logger *Logger) Error(content ...interface{}) {
	if logger.Level&ErrorOn != 0 {
		logger.Log("错误", content...)
	}
}
func (logger *Logger) Fatal(content ...interface{}) {
	if logger.Level&FatalOn != 0 {
		logger.Log("致命", content...)
	}
}

var DefaultLogger *Logger

func init() {
	DefaultLogger = NewLogger(DebugLevel, os.Stdout)
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
