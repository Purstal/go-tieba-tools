package post_finder

import (
	"os"
	"time"

	"github.com/purstal/pbtools/modules/logs"
)

var (
	Logger              *logs.Logger //default
	DelayerLogger       *logs.Logger //
	GettingStructLogger *logs.Logger
)

var logDir string

func InitLoggers() {
	logDir = "log/PostFinder/" + time.Now().Format("20060102_150405") + "/"
	os.MkdirAll(logDir, 0744)

	if logFile, err := os.Create(logDir + "log"); err == nil {
		Logger = logs.NewLoggerWithName("PostFinder", logs.DebugLevel, os.Stdout, logFile)
		Logger.LogWithTime = false
	} else {
		panic("PostFinder.Logger初始化失败,创建log文件失败:" + err.Error())
	}
	if delayer_logFile, err := os.Create(logDir + "log_delayer"); err == nil {
		DelayerLogger = logs.NewLoggerWithName("PostFinder-Delayer", logs.DebugLevel, os.Stdout, delayer_logFile)
	} else {
		Logger.Fatal("PostFinder.DelayerLogger初始化失败,创建log文件失败:" + err.Error())
		panic("PostFinder.DelayerLogger初始化失败,创建log文件失败:" + err.Error())
	}
	if gettingStruct_logFile, err := os.Create(logDir + "log_getting_struct"); err == nil {
		GettingStructLogger = logs.NewLoggerWithName("PostFinder-GettingStruct", logs.DebugLevel, os.Stdout, gettingStruct_logFile)
	} else {
		Logger.Fatal("PostFinder.GettingStructLogger初始化失败,创建log文件失败:" + err.Error())
		panic("PostFinder.GettingStructLogger初始化失败,创建log文件失败:" + err.Error())
	}

}
