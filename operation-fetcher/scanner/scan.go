package scanner

import (
	"fmt"
	"io/ioutil"
	"os"
	//"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/purstal/go-tieba-modules/utils"
)

type DayData struct {
	Time int64
	Logs []Log
}

type MonthDatas struct {
	Datas      []DayData
	hasChanged bool
}

func Scan(BDUSS, forumName string, beginTime, endTime int64, rescan bool) []DayData {

	fmt.Println("开始扫描", beginTime, "-", endTime)

	var scan_begin = time.Now()
	defer func() { fmt.Println("总计扫描耗时", time.Now().Sub(scan_begin)) }()

	var dir = "postlogs/" + forumName + "/"
	var existedFiles = make(map[string]bool)
	if fis, err := ioutil.ReadDir(dir); err != nil {
		os.MkdirAll(dir, 0644)
	} else {
		for _, fi := range fis {
			if !fi.IsDir() {
				existedFiles[fi.Name()] = true
			}
		}
	}

	var datas = make([]DayData, (endTime-beginTime)/(24*60*60)+1)
	//var loadedMonths = make(map[string]*MonthDatas)

	var isFirstTime = true
	var thisMonth time.Time
	var thisMonthDatas *MonthDatas
	var thisMonthFileName string

	for fromTime := beginTime; fromTime < endTime; fromTime += 24 * 60 * 60 {
		var i = int((fromTime - beginTime) / (24 * 60 * 60))
		var goFromTime = time.Unix(fromTime, 0)
		fullFileName := goFromTime.Format("2006-01-02.json")
		if !isSameMonth(thisMonth, goFromTime) {
			if !isFirstTime {
				saveOneMonthDatas(dir, thisMonthFileName, thisMonthDatas)
			} else {
				isFirstTime = false
			}
			thisMonth = goFromTime
			thisMonthFileName = thisMonth.Format("2006-01.tar.gz")

			thisMonthDatas = nil

			if existedFiles[thisMonthFileName] {
				thisMonthDatas = loadMonthDatas(dir, thisMonthFileName, fromTime)
			} else {
				fmt.Println(dir+thisMonthFileName, "本地不存在,创建.")
			}
			if thisMonthDatas == nil {
				if m := int(goFromTime.Month()); m != 2 || isLeap(goFromTime.Year()) {
					thisMonthDatas = &MonthDatas{make([]DayData, COMMON_YEAR_MONTH_DAYS[m]), true}
				} else {
					thisMonthDatas = &MonthDatas{make([]DayData, 29), true}
				}
			}
		}

		var isOk = false
		var day = goFromTime.Day()

		if datas[i] = thisMonthDatas.Datas[day-1]; datas[i].Time != 0 { //exist
			if rescan {
				fmt.Println(dir+thisMonthFileName+"/"+fullFileName, "虽然已存在,但要求重新扫描,重新扫描.")
			} else {
				//fmt.Println(dir+thisMonthFileName+"/"+oldFileName, "已存在,直接读取数据.")
				isOk = true
			}
		} else {
			fmt.Println(dir+thisMonthFileName+"/"+fullFileName, "本地不存在,扫描.")
		}
		if !isOk {
			thisMonthDatas.hasChanged = true
			datas[i].Time = fromTime
			if !isOk {
				scanOneDay_begin := time.Now()
				if BDUSS == "" {
					fmt.Println("没有可用(登陆且有访问后台权限)账号,但需要访问后台,跳过此任务.")
					return nil
				}
				datas[i].Logs = scanOneDay(BDUSS, forumName, fromTime, fromTime+24*60*60-1, 500)
				fmt.Println(goFromTime.Format("2006-01-02"), "扫描耗时", time.Now().Sub(scanOneDay_begin))
			}
			thisMonthDatas.Datas[day-1] = datas[i]
		}

	}

	if !isFirstTime {
		saveOneMonthDatas(dir, thisMonth.Format("2006-01.tar.gz"), thisMonthDatas)
	}

	return datas
	//var logCount = GetLogCount(BDUSS, forumName, time.Unix(beginUnix, 0), time.Unix(beginUnix, 0))
	//println(logCount, beginTime.Unix(), endTime.Unix())
}

func scanOneDay(BDUSS, forumName string, fromTime, toTime int64, maxThreadNumber int) []Log {
	var firstPageDoc *goquery.Document
	var logCount int
	for {
		firstPageDoc = TryGettingListingPostLogDocument(BDUSS, forumName, "", OpType_None, fromTime, toTime, 1)
		var err error
		logCount, err = ExtractLogCount(firstPageDoc)
		if err == nil { //响应的页面异常时提取到的日志数是空字符串,strconv.atoi会返回err
			break
		}
	}

	var logs = make([]Log, logCount)

	if logCount == 0 {
		return nil
	}

	if logCount > 30 {
		if !ExtractLogs(firstPageDoc, logs, 0, 30) {
			TryGettingAndExtractLogs(BDUSS, forumName, "", OpType_None, fromTime, toTime, 1, logs, 30)
		}
	} else {
		if !ExtractLogs(firstPageDoc, logs, 0, logCount) {
			TryGettingAndExtractLogs(BDUSS, forumName, "", OpType_None, fromTime, toTime, 1, logs, logCount)
		}
	}

	var totalPage = (logCount-1)/30 + 1

	if totalPage >= 2 {
		TryGettingAndExtractLogs(BDUSS, forumName, "", OpType_None, fromTime, toTime, totalPage, logs, logCount-(totalPage-1)*30)
	}

	if totalPage > 2 {

		tm := utils.NewLimitTaskManager(maxThreadNumber, totalPage-2)

		//go func() {
		for i := 1; i < totalPage-1; i++ { //第一页最后一页单算
			tm.RequireChan <- true
			if <-tm.DoChan {
				go func(i int) {
					TryGettingAndExtractLogs(BDUSS, forumName, "", OpType_None, fromTime, toTime, i, logs, 30)
					tm.FinishChan <- true
				}(i)

			}
		}
		//}()
		<-tm.AllTaskFinishedChan
	}
	return logs

}

func isLeap(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

func isSameMonth(t1, t2 time.Time) bool {
	if t1.Month() != t2.Month() || t1.Year() != t2.Year() {
		return false
	}
	return true
}

var COMMON_YEAR_MONTH_DAYS = []int{0,
	31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
