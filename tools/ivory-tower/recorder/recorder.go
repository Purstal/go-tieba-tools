package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"time"

	"github.com/BurntSushi/toml"

	"github.com/purstal/pbtools/modules/logs"
	"github.com/purstal/pbtools/modules/postbar"
	"github.com/purstal/pbtools/modules/postbar/apis/forum-win8-1.5.0.0"
	monitors "github.com/purstal/pbtools/tool-cores/forum-page-monitor"
	"github.com/purstal/pbtools/tools/operation-analyser/csv"
)

type Config struct {
	SavePath     string `toml:"save_path"`
	TempFileName string `toml:"temp_file_name"`
	Forum        string `toml:"forum"`
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
	_, err3 := NewRecorder(postbar.NewDefaultWindows8Account(""), config.Forum, config.SavePath, config.TempFileName, logger)
	if err3 != nil {
		return
	}

	<-make(chan struct{})
}

type Recorder struct{}

type TimeDuartion struct {
	Begin *time.Time
	End   *time.Time
}

type Temp struct {
	RecordingDate  int64
	Records        []Record
	RecordsLastDay []Record

	MonitoringTimeLastDay []TimeDuartion
	MonitoringTime        []TimeDuartion
}

func NewRecorder(accWin8 *postbar.Account, kw string,
	savePath, tempFileName string, logger *logs.Logger) (*Recorder, error) {
	var recoder Recorder
	var temp Temp

	os.Mkdir(savePath, 0644)

	if data, err := ioutil.ReadFile(tempFileName); err == nil {
		logger.Info("读取临时文件.")
		json.Unmarshal(data, &temp)
	} else if os.IsNotExist(err) {
		logger.Info("临时文件不存在.")
	} else {
		logger.Fatal("无法打开临时文件.", err)
		return nil, err
	}

	var lastServerTime *time.Time

	{ //处理善后
		var c = make(chan os.Signal)
		go func() {
			s := <-c
			if n := len(temp.MonitoringTimeLastDay); n != 0 {
				temp.MonitoringTimeLastDay[n-1].End = lastServerTime
			} else {
				temp.MonitoringTimeLastDay = append(temp.MonitoringTimeLastDay, TimeDuartion{nil, lastServerTime})
			}
			tempFile, err1 := os.OpenFile(tempFileName, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
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
			tempFile.Close()
			os.Exit(0)
		}()
		signal.Notify(c)
	}

	var serverTimeChan = make(chan time.Time)
	monitor := monitors.NewFreshPostMonitorAdv(accWin8, kw, time.Second, func(page monitors.ForumPage) {
		serverTimeChan <- page.Extra.ServerTime
	})

	var iAmATime int64

	const TIME_DIVIDE = 24 * 60 * 60
	go func() {
		for {
			select {
			case serverTime := <-serverTimeChan:
				if lastServerTime == nil {
					temp.MonitoringTime = append(temp.MonitoringTimeLastDay, TimeDuartion{&serverTime, nil})
				}
				lastServerTime = &serverTime
				if u := serverTime.Unix(); temp.RecordingDate-u >= 24*60*60 {
					temp.MonitoringTimeLastDay = temp.MonitoringTime
					temp.MonitoringTime = []TimeDuartion{TimeDuartion{&serverTime, nil}}
					temp.RecordingDate = u / (24 * 60 * 60) * (24 * 60 * 60)
					temp.RecordsLastDay = temp.Records
					iAmATime = time.Now().Unix()
				}
				if temp.RecordsLastDay != nil && time.Now().Unix()-iAmATime >= 60 {
					{
						t := time.Unix(serverTime.Unix()/(24*60*60)*(24*60*60)-1, 0)
						if n := len(temp.MonitoringTimeLastDay); n != 0 {
							temp.MonitoringTimeLastDay[n-1].End = &t
						} else {
							temp.MonitoringTimeLastDay = append(temp.MonitoringTime, TimeDuartion{nil, &t})
						}
					}
					go func() {
						err := saveRecords(savePath, temp.RecordsLastDay, temp.RecordingDate-24*60*60, temp.MonitoringTimeLastDay)
						if err != nil {
							logs.Fatal("无法保存记录:", err)
						}
						temp.RecordsLastDay = nil
					}()

				}
			case page := <-monitor.PageChan:
				for i := len(page.ThreadList) - 1; i >= 0; i-- {
					if IsNewThread(page.ThreadList[i]) {
						if page.ThreadList[i].LastReplyTime.After(time.Unix(temp.RecordingDate, 0)) {
							temp.Records = append(temp.Records, makeRecord(page.ThreadList[i]))
						} else {
							temp.RecordsLastDay = append(temp.RecordsLastDay, makeRecord(page.ThreadList[i]))
						}
					}
				}
			}
		}
	}()

	return &recoder, nil

}

func IsNewThread(thread *forum.ForumPageThread) bool {
	if thread.ReplyNum == 0 &&
		thread.Author.ID == thread.LastReplyer.ID {
		return true
	}
	return false
}

func saveRecords(path string, records []Record, date int64, monitoringTime []TimeDuartion) error {
	f, err := os.Create(path + time.Unix(date, 0).Format("2006-01-02.csv"))
	if err != nil {
		return err
	}
	w := csv.NewWriter(f)

	var mt_str = "监控时间:["

	if len(monitoringTime) == 0 {
		mt_str += "]"
	} else {
		for i, mt := range monitoringTime {
			mkstr := func(t *time.Time) string {
				if t != nil {
					return t.Format("2006-01-02 15:04:05")
				} else {
					return "未知"
				}
			}
			mt_str += fmt.Sprintf("%s => %s", mkstr(mt.Begin), mkstr(mt.End))
			if i != len(monitoringTime)-1 {
				mt_str += "; "
			} else {
				mt_str += "]"
			}
		}
	}

	w.Write([]string{mt_str})
	w.Write(nil)

	for _, r := range records {
		w.WriteAll([][]string{[]string{fmt.Sprintf("tid: %d", r.Tid), fmt.Sprintf("作者: %s", r.Author),
			fmt.Sprintf("标题: %s", r.Title)}, []string{"时间: %s", time.Unix(r.PostTime, 0).Format("2006-01-02 15:04:05"), r.Abstract}, nil})
	}
	w.Flush()
	f.Close()
	return nil
}

type Record struct {
	Title    string
	Tid      uint64
	PostTime int64
	Author   string
	Abstract string
	IsExist  bool
}

func makeRecord(thread *forum.ForumPageThread) Record {
	return Record{
		Title:    thread.Title,
		Tid:      thread.Tid,
		PostTime: thread.LastReplyTime.Unix(),
		Author:   thread.Author.Name,
		Abstract: extractAbstract(append(thread.Abstract, thread.MediaList...)),
		IsExist:  true,
	}
}

//摘要: [#文字: xxx; #图片: http://xxx]
//这个函数的历史可以追溯到去年,这边复制过来改改继续用
func extractAbstract(contents []interface{}) string {
	if len(contents) == 0 {
		return "摘要: []"
	}
	var str = "摘要: ["

	for i, _content := range contents {
		content, ok1 := _content.(map[string]interface{})
		t, ok2 := content["type"].(string)
		if !ok1 {
			t = "解析失败"
		} else if !ok2 {
			t = "缺少类型"
		}

		switch t {
		case "0": //文字
			if content["text"] != nil {
				str += fmt.Sprintf("#文字: %s", fmt.Sprint(content["text"]))
			} else {
				str += fmt.Sprintf("#文字: 解析失败(%s)", fmt.Sprint(content))
			}
		case "3": //图片
			if content["big_pic"] != nil {
				str += fmt.Sprintf("#图片: %s", fmt.Sprint(content["big_pic"]))
			} else {
				str += fmt.Sprintf("#图片: 解析失败(%s)", fmt.Sprint(content))
			}
		case "5": //视频
			if content["vhsrc"] != nil {
				str += fmt.Sprintf("#视频: %s", fmt.Sprint(content["vhsrc"]))
				if content["vsrc"] != nil {
					str += fmt.Sprint("(%s)", content["vsrc"])
				}
			} else if content["vsrc"] != nil {
				str += fmt.Sprintf("#视频: %s", fmt.Sprint(content["vsrc"]))
			} else {
				str += fmt.Sprintf("#视频: 解析失败(%s)", fmt.Sprint(content))
			}
		case "6":
			if content["src"] != nil {
				str += fmt.Sprintf("#音乐: %s", fmt.Sprint(content["src"]))
			} else {
				str += fmt.Sprintf("#音乐: 解析失败(%s)", fmt.Sprint(content))
			}
		case "解析失败":
			str += fmt.Sprintf("#未知: 解析失败(%s)", fmt.Sprint(content))
		case "缺少类型":
			str += fmt.Sprintf("#未知: 缺少类型(%s)", fmt.Sprint(content))
		default:
			str += fmt.Sprintf("#未知: 未知类型(%s)(%s)", t, fmt.Sprint(content))
		}
		if i != len(contents)-1 {
			str += "; "
		} else {
			str += "]"
		}
	}

	return str
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
