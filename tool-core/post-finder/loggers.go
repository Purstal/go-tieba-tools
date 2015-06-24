package post_finder

import (
	"os"

	"github.com/purstal/pbtools/modules/logs"
)

var logger, delayerLogger, unmarshallerLogger *logs.Logger

func initLoggers(pf *PostFinder, logDir string) (ok bool) {
	os.MkdirAll(logDir, 0644)

	initFunc := func(loggerName string) *logs.Logger {
		logFile, err := os.Create(logDir + loggerName)
		if err != nil {
			logs.Fatal("无法创建log文件.", err)
			return nil
		}
		return logs.NewLogger(logs.DebugLevel, os.Stdout, logFile)
	}

	var loggers [4]*logs.Logger

	for i, loggerName := range []string{
		"post-finder-日志.log",
		"post-finder-延时搜索日志.log",
		"post-finder-解组错误记录.log",
	} {
		loggers[i] = initFunc(loggerName)
		if loggers[i] == nil {
			return false
		}
	}

	logger = loggers[0]
	delayerLogger = loggers[1]
	unmarshallerLogger = loggers[2]

	return true

}
