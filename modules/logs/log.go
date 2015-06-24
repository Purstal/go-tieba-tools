package logs

import (
	"fmt"
	"io"
	"sync"
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
	Name string

	Writers []io.Writer

	Level uint8

	LogWithTime bool

	LastLogTime time.Time

	ColorfulConsoleOutput bool

	Lock sync.Mutex
}

func NewLogger(level uint8, w ...io.Writer) *Logger {
	return &Logger{
		Writers:               w,
		Level:                 level,
		LogWithTime:           true,
		ColorfulConsoleOutput: true,
	}
}

func NewLoggerWithName(name string, level uint8, w ...io.Writer) *Logger {
	return &Logger{
		Name:                  name,
		Writers:               w,
		Level:                 level,
		LogWithTime:           true,
		ColorfulConsoleOutput: true,
	}
}

func (logger *Logger) Log(prefix string, content ...interface{}) {
	logger.Lock.Lock()
	logger.LogTime()
	for _, writer := range logger.Writers {
		log(writer, logger.Name, prefix, content...)
	}
	logger.Lock.Unlock()
}

func log(writer io.Writer, loggerName, prefix string, content ...interface{}) {
	io.WriteString(writer, loggerName+"#"+prefix+"# "+fmt.Sprintln(content...))
}

func (logger *Logger) LogTime() {
	now := time.Now()

	if logger.LastLogTime.Day() != now.Day() {
		for _, writer := range logger.Writers {
			if logger.ColorfulConsoleOutput {
				logColorful(writer, logger.Name, "日期", "36;1", now.Format("L2006-01-02"))
			} else {
				log(writer, logger.Name, "日期", now.Format("L2006-01-02"))
			}
		}
	}
	if logger.LogWithTime {
		if logger.LastLogTime.Unix() != now.Unix() {
			for _, writer := range logger.Writers {
				if logger.ColorfulConsoleOutput {
					logColorful(writer, logger.Name, "时间", "36;1", now.Format("L15:04:05"))
				} else {
					log(writer, logger.Name, "时间", now.Format("L15:04:05"))
				}
			}
		}
	}

	logger.LastLogTime = now
}

func (logger *Logger) Debug(content ...interface{}) {
	if logger.Level&DebugOn != 0 {
		if logger.ColorfulConsoleOutput {
			logger.LogColorful("调试", "32;1", content...)
		} else {
			logger.Log("调试", content...)
		}
	}
}
func (logger *Logger) Info(content ...interface{}) {
	if logger.Level&InfoOn != 0 {
		if logger.ColorfulConsoleOutput {
			logger.LogColorful("信息", "34;1", content...)
		} else {
			logger.Log("信息", content...)
		}
	}
}
func (logger *Logger) Warn(content ...interface{}) {
	if logger.Level&WarnOn != 0 {
		if logger.ColorfulConsoleOutput {
			logger.LogColorful("警告", "33;1", content...)
		} else {
			logger.Log("警告", content...)
		}
	}
}
func (logger *Logger) Error(content ...interface{}) {
	if logger.Level&ErrorOn != 0 {
		if logger.ColorfulConsoleOutput {
			logger.LogColorful("错误", "31;1", content...)
		} else {
			logger.Log("错误", content...)
		}
	}
}
func (logger *Logger) Fatal(content ...interface{}) {
	if logger.Level&FatalOn != 0 {
		if logger.ColorfulConsoleOutput {
			logger.LogColorful("致命", "35;1", content...)
		} else {
			logger.Log("致命", content...)
		}
	}
}
