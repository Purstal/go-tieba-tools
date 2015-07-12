package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/signal"
	"time"

	"github.com/BurntSushi/toml"

	"github.com/purstal/pbtools/modules/logs"
	"github.com/purstal/pbtools/modules/postbar"
	"github.com/purstal/pbtools/modules/postbar/apis/forum-win8-1.5.0.0"
	monitors "github.com/purstal/pbtools/tool-cores/forum-page-monitor"
)

type Record struct {
	Title    string
	Tid      uint64
	PostTime uint64
	Author   string
	Abstract string
	IsExist  bool
}

func main() {
	config, err1 := readConfig()
	if err1 != nil {
		panic(err1)
	}

	os.MkdirAll(config.SavePath, 0644)
	os.Mkdir("logs", 0644)
	logFile, err2 := os.Create("logs/" + time.Now().Format("2006-01-02 15-04-05.log"))
	if err2 != nil {
		panic(err2)
	}
	logger := logs.NewLogger(logs.DebugLevel, os.Stdout, logFile)

}

type Config struct {
	SavePath     string `toml:"save_path"`
	TempFileName string `toml:"temp_file_name"`
}

type Recorder struct{}

type Temp struct {
	RecordingDate  time.Time
	Records        []Record
	RecordsLastDay []Record

	MonitoringTime []struct {
		Begin time.Time
		End   time.Time
	}
}

func NewRecorder(accWin8 *postbar.Account, kw string,
	savePath, tempFileName string, logger *logs.Logger) (*Recorder, error) {
	var recoder Recorder
	var temp Temp

	if data, err := ioutil.ReadFile(tempFileName); err == nil {
		logger.Info("读取临时文件.")
		json.Unmarshal(data, &temp)
	} else if os.IsNotExist(err) {
		logger.Info("临时文件不存在.")
	} else {
		logger.Fatal("无法打开临时文件.", err)
		return nil, err
	}

	{ //处理善后
		var c chan os.Signal
		go func() {
			s := <-c
			tempFile, err1 := os.OpenFile(tempFileName, os.O_TRUNC, 0644)
			logger.Warn("程序结束运行:", s.String())
			if err1 != nil {
				logger.Fatal("无法打开临时文件:", err1)
				os.Exit(1)
			}
			data, _ := json.Marshal(temp)
			_, err2 := tempFile.Write(data)
			if err2 != nil {
				logger.Fatal("无法写入临时文件:", err2)
				os.Exit(1)
			}
			logger.Info("成功保存临时文件.")
			os.Exit(0)
		}()
		signal.Notify(c)
	}

	var serverTimeChan = make(chan time.Time)
	monitor := monitors.NewFreshPostMonitorAdv(accWin8, kw, time.Second, func(page monitors.ForumPage) {
		serverTimeChan <- page.Extra.ServerTime
	})

	go func() {
		for {
			select {
			case serverTime := <-serverTimeChan:

			case page := <-monitor.PageChan:
			}
		}
	}()

}

func readConfig() (*Config, error) {
	data, err := ioutil.ReadFile(`config.toml`)
	if err != nil {
		return nil, err
	}
	var config Config
	err = toml.Unmarshal(data, &config)
	return &config, err
}
