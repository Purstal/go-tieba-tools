package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"

	"github.com/purstal/pbtools/modules/postbar"
	"github.com/purstal/pbtools/modules/postbar/apis"

	"github.com/purstal/pbtools/modules/logs"
	oldlog "github.com/purstal/pbtools/tool-cores/operation-analyser/old/log"

	old "github.com/purstal/pbtools/tool-cores/operation-analyser/old/caozuoliang"
	"github.com/purstal/pbtools/tool-cores/operation-analyser/old/inireader"

	analyser "github.com/purstal/pbtools/tool-cores/operation-analyser"
	myanalyse "github.com/purstal/pbtools/tools/operation-analyser/analyse"
)

func analyse(datas []analyser.DayData) {
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

	var useOldVersionPtr = flag.Bool(`use-old-version`, false, `使用1.6.2版本的方式(old.ini)扫描/统计.`)
	var onlyScanPtr = flag.Bool(`only-scan`, false, `只扫描,对use-old-version无效.`)
	var rescan = flag.Bool(`rescan`, false, `重新扫描,即使扫描过.`)
	var taskFileNamePtr = flag.String(`task-file`, "", `指定任务文件,需带上扩展名,对use-old-version有效.`)

	flag.Parse()

	main_begin := time.Now()

	////////////////////////////////////////////////////////////////开始初始化Logger////////////////////////////////////////////////////////////////

	var logDir = "log/operation-analyser/"
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

	if *useOldVersionPtr {
		logs.Warn("旧版本将使用旧版日志系统.")
		oldlog.INIT_LOG(config.Log.Log1, config.Log.Log2, config.Log.Log3)
	}

	logs.Info("百度贴吧操作量统计工具 by purstal " + old.Version)

	var config_json, _ = json.Marshal(config)
	logs.Info("flag =", os.Args[1:])
	logs.Info("config =", string(config_json))

	if *useOldVersionPtr {
		old.INIT(config.Net.MaxRetryTime,
			time.Millisecond*time.Duration(config.Net.TimeOutMs),
			/*旧贴*/ 720,
			/*同账号*/ 2,
			/*同主题*/ 2)
	}

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

	if *useOldVersionPtr {
		if *taskFileNamePtr != "" {
			oldVersion(*taskFileNamePtr, accountMap, &config)
		} else {
			oldVersion("old.ini", accountMap, &config)
		}
	} else {
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

			var datas = analyser.Scan(BDUSS, task.ForumName, beginDate.Unix(), endDate.Unix()+24*60*60-1, *rescan)

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

	logs.Info("百度贴吧操作量统计工具 by purstal " + old.Version)

}

func ReadFile(fileName string) ([]byte, error) {
	var data, err = ioutil.ReadFile(fileName)

	if err != nil {
		return data, err
	}

	if len(data) < 2 {
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

func oldVersion(taskFileName string, accountMap map[string]*Account, config *Config) {
	var 贴吧, err3 = inireader.ReadINI(taskFileName)
	if err3 != nil {
		oldlog.Loglog("读取设置文件失败:贴吧")
		oldlog.Loglog(err3.Error())

	}

	for key, value := range 贴吧 {

		if value["停用"] == "true" {
			oldlog.Loglog("根据\"贴吧.ini\"中的设置,跳过", key, "吧")
			continue
		}

		bt_pb := time.Now()
		oldlog.Loglog("开始统计", key, "吧")

		var bt, et time.Time

		var by, bm, bd int
		i, _ := fmt.Sscanf(value["开始时间"], "%d-%d-%d", &by, &bm, &bd)
		if i != 3 {
			if value["开始时间"] == "" {
				oldlog.Loglog("开始时间省缺,使用1970-1-1")
			} else {
				oldlog.Loglog("开始时间输入格式有误,使用1970-1-1")
				oldlog.Loglog(value["开始时间"])
			}

			bt = time.Date(1970, 1, 1, 0, 0, 0, 0, time.Local)
		} else {
			bt = time.Date(by, time.Month(bm), bd, 0, 0, 0, 0, time.Local)
		}
		oldlog.Loglog("开始时间:", bt.String())

		var ey, em, ed int
		j, _ := fmt.Sscanf(value["结束时间"], "%d-%d-%d", &ey, &em, &ed)
		if j != 3 {
			if value["结束时间"] == "" {
				oldlog.Loglog("结束时间省缺,使用昨天的日期")
			} else {
				oldlog.Loglog("结束时间输入格式有误,使用昨天的日期")
				oldlog.Loglog(value["结束时间"])
			}
			tn := time.Now()
			et = time.Date(tn.Year(), tn.Month(), tn.Day(), 0, 0, 0, 0, time.Local)
		} else {
			et = time.Date(ey, time.Month(em), ed, 0, 0, 0, 0, time.Local)
		}
		oldlog.Loglog("结束时间:", et.String())

		whitelist := make(map[string]bool)
		aaaa := value["白名"]
		oldlog.Loglog(key, "吧吧务白名:", aaaa)
		for _, white := range strings.Split(aaaa, ",") {
			whitelist[strings.ToLower(white)] = true
		}

		所需账号字符串 := value["账号"]
		针对 := value["针对"]
		var BDUSS string
		var _bawulist []old.Bawu
		var 所需账号切片 []string

		所需账号切片 = strings.Split(所需账号字符串, ",")
		for _, 账号 := range 所需账号切片 {
			if BDUSS = accountMap[账号].BDUSS; BDUSS != "" {
				oldlog.Loglog("尝试通过账号", 账号, "获取", key, "吧吧务名单,并测试能否访问吧务后台")
				_bawulist = old.GetBawuList(BDUSS, key, whitelist)
				if len(_bawulist) == 0 {
					oldlog.Loglog(账号, "未能获取", key, "吧吧务名单,无法访问吧务后台,放弃")
					continue
				}
				oldlog.Loglog(账号, "成功获取", key, "吧吧务名单,可以访问吧务后台")

				if 针对 != "" {
					oldlog.Loglog("已设置针对的吧务,仅对针对的吧务进行统计")
					针对切片 := strings.Split(针对, ",")
					oldlog.Loglog("针对:", 针对)
					__bawulist := old.GetBawuList_C(针对切片)

					for i, __bawu := range __bawulist {
						find := false
						for _, _bawu := range _bawulist {
							if strings.ToLower(_bawu.Username) == strings.ToLower(__bawu.Username) {
								__bawulist[i] = _bawu
								find = true
							}
						}
						if !find {
							oldlog.Loglog("注意:吧务", __bawu.Username, "已经离开吧务团队.其文件名(大小写相关)将依据设置文件中的设置取名")
						}
					}
					_bawulist = __bawulist

				}
				var 杂项切片 []string
				if 杂项 := value["杂项"]; 杂项 != "" {
					oldlog.Loglog("已设置统计杂项")
					oldlog.Loglog("杂项:", 杂项)
					杂项切片 = strings.Split(杂项, ",")
				}

				var 整体比较 bool
				var 整体比较标准线 int = 100
				var zwl = value["整体比较白名"]
				var zwlm map[string]bool
				if value["整体比较"] == "true" {
					oldlog.Loglog("已开启整体比较")
					整体比较 = true

					if z := value["整体比较标准线"]; z != "" {
						oldlog.Loglog("已设置整体比较")
						if 标准线, err := strconv.Atoi(value["整体比较标准线"]); err != nil {
							oldlog.Loglog("整体比较输入格式有误,使用", 100)
						} else {
							整体比较标准线 = 标准线
						}
					}

					oldlog.Loglog(key, "吧整体比较白名:", zwl)
					if zwls := strings.Split(zwl, ","); len(zwls) != 0 {
						zwlm = make(map[string]bool)
						for _, white := range zwls {
							zwlm[strings.ToLower(white)] = true
						}
					}
				}

				old.Do(BDUSS, key, _bawulist, 杂项切片, 整体比较, 整体比较标准线, zwlm, &bt, &et,
					config.Thread.ScanGoroutineNumber, config.Thread.AnalyseGoroutineNumberPerScanGorountine)
				break
			}
		}

		if len(_bawulist) == 0 {
			oldlog.Loglog("跳过", key, "吧.没有账号能够访问吧务后台")
			continue
		}

		oldlog.Loglog("完成统计", key, "吧", "用时", time.Now().Sub(bt_pb).String())

	}

}

func findAvailableBDUSS(forumName string, accountMap map[string]*Account, accounts []string) string {
	for _, account := range accounts {
		if accountMap[account] != nil && accountMap[account].BDUSS != "" {
			var doc = analyser.TryGettingListingPostLogDocument(accountMap[account].BDUSS, forumName, "", analyser.OpType_None, 0, 0, 1)
			if analyser.CanViewBackstage(doc) {
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
