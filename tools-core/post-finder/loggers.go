package post_finder

import (
	"os"
	"strconv"
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
	logDir = "log/PostFinder/" + strconv.FormatInt(time.Now().Unix(), 16) + "/"
	os.MkdirAll(logDir, 0744)

	if logFile, err := os.Create(logDir + "log"); err == nil {
		Logger = logs.NewLogger(logs.DebugLevel, os.Stdout, logFile)
		Logger.LogWithTime = false
	} else {
		panic("删贴机Logger初始化失败,创建log文件失败:" + err.Error())
	}
	if delayer_logFile, err := os.Create(logDir + "log_delayer"); err == nil {
		DelayerLogger = logs.NewLogger(logs.DebugLevel, os.Stdout, delayer_logFile)
	} else {
		Logger.Fatal("删贴机DelayerLogger初始化失败,创建log文件失败:" + err.Error())
		panic("删贴机DelayerLogger初始化失败,创建log文件失败:" + err.Error())
	}
	if gettingStruct_logFile, err := os.Create(logDir + "log_getting_struct"); err == nil {
		GettingStructLogger = logs.NewLogger(logs.DebugLevel, os.Stdout, gettingStruct_logFile)
	} else {
		Logger.Fatal("删贴机GettingStructLogger初始化失败,创建log文件失败:" + err.Error())
		panic("删贴机GettingStructLogger初始化失败,创建log文件失败:" + err.Error())
	}

}
