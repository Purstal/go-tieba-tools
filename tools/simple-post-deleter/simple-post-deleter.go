package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/purstal/pbtools/modules/logs"
	"github.com/purstal/pbtools/modules/postbar"
	"github.com/purstal/pbtools/modules/postbar/apis"

	postdeleter "github.com/purstal/pbtools/tools/simple-post-deleter/post-deleter"
)

var accWin8 *postbar.Account
var accAndr *postbar.Account

func init() {
	go func() {
		http.ListenAndServe(":33101", nil)
	}()
}

type Settings struct {
	BDUSS     string
	ForumName string
	ForumID   uint64

	ContentRegexpFilePath     string `json:"贴子内容正则文件"`
	UserNameRegexpFilePath    string `json:"用户名正则文件"`
	TidWhiteListFilePath      string `json:"tid白名单文件"`
	UserNameWhiteListFilePath string `json:"用户名白名单文件"`
	BawuListFilePath          string `json:"吧务列表文件"`
}

var settings *Settings

func main() {
	os.MkdirAll("log/simplePostDeleter", 0644)
	logfile, err1 := os.Create("log/simplePostDeleter/" + time.Now().Format("20060102-150405-通常.log"))
	if err1 != nil {
		logs.Fatal("无法创建log文件.", err1)
	} else {
		defer logfile.Close()
	}

	logger := logs.NewLogger(logs.DebugLevel, os.Stdout, logfile)
	logs.SetDefaultLogger(logger)
	logs.DefaultLogger.LogWithTime = false
	logs.Info("删贴机启动", time.Now())

	keepUpdatingSettings()
	{
		var BDUSS = settings.BDUSS
		if BDUSS != "" {
			settings.BDUSS = "***"
		}
		logs.Info(settings)
		settings.BDUSS = BDUSS
	}

	if settings == nil {
		os.Exit(1)
	}

	if settings.BDUSS == "" {
		logs.Warn("未设置BDUSS.")
	}

	var accAndr = postbar.NewDefaultAndroidAccount("")
	var accWin8 = postbar.NewDefaultWindows8Account("")
	accWin8.BDUSS = settings.BDUSS
	accAndr.BDUSS = settings.BDUSS

	os.MkdirAll("log/simplePostDeleter", 0644)
	opLogfile, err2 := os.Create("log/simplePostDeleter/" + time.Now().Format("20060102-150405-操作.log"))
	if err2 != nil {
		logs.Fatal("无法创建log文件.", err2)
	} else {
		defer logfile.Close()
	}

	if settings.ForumID == 0 {
		logs.Info("设置中未提供ForumID,自动获取.")
		settings.ForumID = getFid(settings.ForumName)
		if settings.ForumID == 0 {
			logs.Fatal("未能获得到fid,退出.")
			os.Exit(1)
		}
	}

	opLogger := logs.NewLogger(logs.DebugLevel, os.Stdout, opLogfile)

	d := postdeleter.NewPostDeleter(accWin8, accAndr, settings.ForumName, settings.ForumID,
		settings.ContentRegexpFilePath, settings.UserNameRegexpFilePath,
		settings.TidWhiteListFilePath, settings.UserNameWhiteListFilePath, settings.BawuListFilePath,
		logger, opLogger)

	if d == nil {
		logs.Fatal("无法启动删贴机,退出.")
		os.Exit(1)
	}

	d.Run(time.Second)

	<-make(chan struct{})

}

func getFid(forumName string) uint64 {
	var fid uint64

	for i := 0; ; {
		_fid, err, pberr := apis.GetFid(forumName)
		if err != nil {
			continue
		} else if pberr != nil && pberr.ErrorCode != 0 {
			if i < 10 {
				i++
				continue
			}
			logs.Fatal("方案A未能获得到fid,进行下一步尝试.", pberr)
			break
		} else if _fid == 0 {
			logs.Warn("方案A未能获得到fid,进行下一步尝试.")
			break
		} else {
			fid = _fid
			break
		}
	}

	if fid == 0 {
		for i := 0; ; {
			results, err, pberr := apis.SearchForum(forumName)
			if err != nil {
				continue
			} else if pberr != nil && pberr.ErrorCode != 0 {
				if i < 10 {
					i++
					continue
				}
				logs.Fatal("方案B未能获得到fid,放弃.", pberr)
				return 0
			} else if len(results) == 0 {
				logs.Fatal("未找到该贴吧,放弃.", pberr)
				return 0
			} else {
				for _, result := range results {
					if result.ForumName == forumName {
						return result.ForumID
					}
				}
				logs.Fatal("未找到该贴吧,放弃.", pberr)
				return 0
			}
		}
	}

	return fid
}

func useless() {
	fmt.Println(io.EOF,
		http.DefaultMaxHeaderBytes,
	)
}

func LoadSettings(fileName string) (*Settings, error) {

	file, err := os.Open(fileName)

	if err != nil {
		return nil, err
	}

	data, err2 := ioutil.ReadAll(file)
	if err2 != nil {
		return nil, err2
	}
	if data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		data = data[3:]
	}

	var settings Settings

	err3 := json.Unmarshal(data, &settings)
	if err3 != nil {
		return nil, err3
	}
	return &settings, nil
}

func keepUpdatingSettings() {
	var fileName string
	if len(os.Args) == 1 {
		fileName = "删贴机设置.json"
	} else {
		fileName = os.Args[1]
	}

	var lastModTime time.Time
	ticker := time.NewTicker(time.Second)
	var isFirstTime bool = true
	var firstTimeWaitChan = make(chan bool)
	go func() {
		for {

			info, err := os.Stat(fileName)
			if err != nil {
				if isFirstTime {
					panic(err)
				}
				continue
			}

			if modTime := info.ModTime(); modTime.After(lastModTime) {
				lastModTime = modTime
				_settings, err := LoadSettings(fileName)
				if err != nil {
					logs.Fatal("更新设置文件失败,将继续使用旧设置:", err)
				} else {
					logs.Info("更新设置文件成功(然而这并没有什么用(除了第一次之外)).")
					settings = _settings
				}
			}

			if isFirstTime {
				firstTimeWaitChan <- true
				isFirstTime = false
			}

			<-ticker.C
		}

	}()

	<-firstTimeWaitChan
	close(firstTimeWaitChan)
}
