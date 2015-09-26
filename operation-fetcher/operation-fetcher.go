package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"
	//"strconv"
	//"strings"

	"github.com/BurntSushi/toml"

	"github.com/purstal/go-tieba-base/tieba"
	"github.com/purstal/go-tieba-base/tieba/apis"

	"github.com/purstal/go-tieba-base/logs"

	//oldlog "github.com/purstal/go-tieba-modules/operation-scanner/old/log"
	//old "github.com/purstal/go-tieba-modules/operation-scanner/old/caozuoliang"
	//"github.com/purstal/go-tieba-modules/operation-scanner/old/inireader"

	myanalyse "github.com/purstal/go-tieba-tools/operation-fetcher/analyse"
	"github.com/purstal/go-tieba-tools/operation-fetcher/scanner"
)

const VERSION = "2.1.0"

func analyse(datas []scanner.DayData) {
	myanalyse.Analyse2(datas)
}

func main() {
	_main()
	//oldlog.Loglog("程序运行完成,按回车键退出")
	//bufio.NewReader(os.Stdin).ReadLine()
}

type Config struct {
	Thread struct {
		ScanGoroutineNumber                     int
		AnalyseGoroutineNumberPerScanGorountine int
	}
	Net struct {
		MaxRetryTime int
		TimeOutMs    int
	}
	Log struct {
		Log1, //旧称loglog
		Log2, //旧称log233
		Log3 bool //旧称log!!!
	}
}

type Account struct {
	UserName string
	Password string
	BDUSS    string
}

type Yardstick struct {
	ReportDeletingOldPostDay   int `toml:"旧贴日数"`
	ReportDeletingSameUser     int `toml:"报告同账号"`
	ReportDeletingInSameThread int `toml:"报告同主题"`
}

type Task struct {
	Accounts           []string
	ForumName          string
	BeginDate, EndDate string
}

func _main() {

	//var useOldVersionPtr = flag.Bool(`use-old-version`, false, `使用1.6.2版本的方式(old.ini)扫描/统计.`)
	var onlyScanPtr = flag.Bool(`only-scan`, false, `只扫描,对use-old-version无效.`)
	var rescan = flag.Bool(`rescan`, false, `重新扫描,即使扫描过.`)
	var taskFileNamePtr = flag.String(`task-file`, "", `指定任务文件,需带上扩展名,对use-old-version有效.`)

	flag.Parse()

	main_begin := time.Now()

	////////////////////////////////////////////////////////////////开始初始化Logger////////////////////////////////////////////////////////////////

	var logDir = "log/operation-scanner/"
	if err := os.MkdirAll(logDir, 0644); err != nil {
		logs.Error("无法创建log目录,将不保存日志(可以尝试使用重定向保存日志).", err)

	} else {
		if logfile, err := os.Create(logDir + time.Now().Format("20060102-150405.log")); err != nil {
			logs.Error("无法创建log文件,将不保存日志(可以尝试使用重定向保存日志).", err)
		} else {
			logs.SetDefaultLogger(logs.NewLogger(logs.DebugLevel, os.Stdout, logfile))
		}
	}

	////////////////////////////////////////////////////////////////结束初始化Logger////////////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////////////开始设置细节////////////////////////////////////////////////////////////////

	var configSrc, err1 = ReadFile("config.toml")
	if err1 != nil {
		fmt.Println(err1.Error())
		return
	}
	var config Config
	if _, err := toml.Decode(string(configSrc), &config); err != nil {
		logs.Fatal()
		return
	}

	/*
		if *useOldVersionPtr {
			logs.Warn("旧版本将使用旧版日志系统.")
			oldlog.INIT_LOG(config.Log.Log1, config.Log.Log2, config.Log.Log3)
		}
	*/

	logs.Info("百度贴吧操作量统计工具 by purstal " + VERSION)

	var config_json, _ = json.Marshal(config)
	logs.Info("flag =", os.Args[1:])
	logs.Info("config =", string(config_json))

	/*
		if *useOldVersionPtr {
			old.INIT(config.Net.MaxRetryTime,
				time.Millisecond*time.Duration(config.Net.TimeOutMs),
				720, //<-旧贴
				2,   //<-同账号
				2)   //<-同主题
		}
	*/

	////////////////////////////////////////////////////////////////结束设置细节////////////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////////////开始设置账号////////////////////////////////////////////////////////////////
	var accountSrc, _ = ReadFile("account.toml")
	var accountMap map[string]*Account
	if _, err := toml.Decode(string(accountSrc), &accountMap); err != nil {
		logs.Fatal("account.toml", err.Error())
		return
	}
	checkAccount(accountMap)

	////////////////////////////////////////////////////////////////结束设置账号////////////////////////////////////////////////////////////////
	/*
		if *useOldVersionPtr {
			if *taskFileNamePtr != "" {
				oldVersion(*taskFileNamePtr, accountMap, &config)
			} else {
				oldVersion("old.ini", accountMap, &config)
			}
		} else */
	{
		var taskFileName string
		if taskFileName = *taskFileNamePtr; taskFileName == "" {
			taskFileName = "task.toml"
		}
		var taskSrc, _ = ReadFile(taskFileName)
		var tasks struct {
			Task []Task
		}
		if _, err := toml.Decode(string(taskSrc), &tasks); err != nil {
			logs.Fatal(taskFileName, err.Error())
			return
		}

		for _, task := range tasks.Task {
			task_begin := time.Now()
			var task_json, _ = json.Marshal(task)
			logs.Info("任务 =", string(task_json))
			var BDUSS = findAvailableBDUSS(task.ForumName, accountMap, task.Accounts)
			if BDUSS == "" {
				//oldlog.Loglog("由于没有可用(登陆且有访问后台权限)账号,跳过此任务.")
				logs.Warn("没有可用(登陆且有访问后台权限)账号,如果需要访问后台则会跳过此任务.")
				//continue
			}

			var beginDate, err1 = parseDate(task.BeginDate)
			if err1 != nil {
				logs.Fatal("解析task.BeginDate失败,跳过任务.", err1)
				continue
			}
			var endDate, err2 = parseDate(task.EndDate)
			if err2 != nil {
				logs.Fatal("解析task.EndDate失败,跳过任务.", err2)
				continue
			}

			var datas = scanner.Scan(BDUSS, task.ForumName, beginDate.Unix(), endDate.Unix()+24*60*60-1, *rescan)

			if datas == nil {
				logs.Fatal("扫描失败,本任务失败,用时", time.Now().Sub(task_begin))
			} else if *onlyScanPtr {
				logs.Info("扫描完成,本任务完成,用时", time.Now().Sub(task_begin))
			} else {
				logs.Info("扫描完成,用时", time.Now().Sub(task_begin))

				analyse_begin := time.Now()
				analyse(datas)
				logs.Info("分析完成,本任务完成,用时", time.Now().Sub(analyse_begin))

			}
		}
	}

	main_end := time.Now()
	main_diration := main_end.Sub(main_begin).String()
	logs.Info("用时", main_diration)

	logs.Info("百度贴吧操作量统计工具 by purstal " + VERSION)

}

func ReadFile(fileName string) ([]byte, error) {
	var data, err = ioutil.ReadFile(fileName)

	if err != nil {
		return data, err
	}

	if len(data) < 3 {
		return data, nil
	}
	if data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {

		return data[3:], nil
	}
	return data, nil
}

func checkAccount(accountMap map[string]*Account) {

	for userName, value := range accountMap {
	RETRY:
		if value.BDUSS != "" {
			if isLogin, err := apis.IsLogin(value.BDUSS); err == nil && isLogin {
				logs.Info("用户", userName, "BDUSS验证成功")
				continue
			} else {
				logs.Info("用户", userName, "BDUSS已失效,尝试使用正常方式登陆")
				value.BDUSS = ""
				goto RETRY
			}
		} else if value.Password == "" {
			logs.Warn("用户", userName, "没有有效的BDUSS,且没有设置密码,登录失败")
		} else {
			var acc = postbar.NewDefaultWindows8Account(userName)
			var err, pberr = apis.Login(acc, value.Password)
			if err != nil || (pberr != nil && pberr.ErrorCode != 0) {
				logs.Warn("用户", userName, "登录失败:", err.Error(), pberr)
			} else {
				value.BDUSS = acc.BDUSS
				logs.Info("用户", userName, "登录成功")
			}
		}
	}
}

func findAvailableBDUSS(forumName string, accountMap map[string]*Account, accounts []string) string {
	for _, account := range accounts {
		if accountMap[account] != nil && accountMap[account].BDUSS != "" {
			var doc = scanner.TryGettingListingPostLogDocument(accountMap[account].BDUSS, forumName, "", scanner.OpType_None, 0, 0, 1)
			if scanner.CanViewBackstage(doc) {
				return accountMap[account].BDUSS
			}
		}
	}
	return ""
}

func parseDate(str string) (*time.Time, error) {
	var year, month, day int
	var _, err = fmt.Sscanf(str, "%d-%d-%d", &year, &month, &day)
	if err != nil {
		return nil, err
	}
	var t = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
	return &t, nil
}
