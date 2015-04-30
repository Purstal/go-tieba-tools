package scaner

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	//"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type DayData struct {
	Time int64
	Logs []Log
}

type YearDatas struct {
	Datas      []DayData
	hasChanged bool
}

func Scan(BDUSS, forumName string, beginTime, endTime int64, rescan bool) []DayData {
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
	var loadedYears = make(map[string]*YearDatas)

	for fromTime := beginTime; fromTime < endTime; fromTime += 24 * 60 * 60 {
		var i = int((fromTime - beginTime) / (24 * 60 * 60))
		//datas[i].Time = fromTime
		oldFileName := time.Unix(fromTime, 0).Format("2006-01-02.json")
		oldFileName2 := time.Unix(fromTime, 0).Format("2006-01-02.json.tar.gz")
		fileName := time.Unix(fromTime, 0).Format("2006.tar.gz")
		var isOk, isOld, isOld2 = false, false, false
		var day = time.Unix(fromTime, 0).YearDay()
		var yearDatas *YearDatas
		{
			yearDatas = loadedYears[fileName]
			if yearDatas == nil {
				if existedFiles[fileName] {
					yearDatas = loadYearDatas(dir, fileName, fromTime)
				} else {
					fmt.Println(dir+fileName, "不存在,创建.")
				}
				if yearDatas == nil {
					if isLeap(time.Unix(fromTime, 0).Year()) {
						yearDatas = &YearDatas{make([]DayData, 366), true}
					} else {
						yearDatas = &YearDatas{make([]DayData, 365), true}
					}
				}
				loadedYears[fileName] = yearDatas
			}

		}
		//fmt.Println(yearDatas.Datas[day-1])
		if datas[i] = yearDatas.Datas[day-1]; datas[i].Time != 0 { //exist
			if rescan {
				fmt.Println(dir+fileName+"/"+oldFileName, "虽然已存在,但要求重新扫描,重新扫描.")
			} else {
				//fmt.Println(dir+fileName+"/"+oldFileName, "已存在,直接读取数据.")
				isOk = true
			}
		} else {
			fmt.Println(dir+fileName+"/"+oldFileName, "本地不存在,扫描.")
		}
		if existedFiles[oldFileName2] && !isOk {
			isOld2 = true
			isOk = loadData(datas, i, BDUSS, forumName, fromTime, dir, oldFileName2, rescan)
		}
		if existedFiles[oldFileName] && !isOk {
			isOld = true
			isOk = loadOldData(datas, i, BDUSS, forumName, fromTime, dir, oldFileName, rescan)
		}
		if !isOk || isOld || isOld2 {
			yearDatas.hasChanged = true
			datas[i].Time = fromTime
			if !isOk {
				scanOneDay_begin := time.Now()
				datas[i].Logs = scanOneDay(BDUSS, forumName, fromTime, fromTime+24*60*60-1, 500)
				fmt.Println(time.Unix(fromTime, 0).Format("2006-01-02"), "扫描耗时", time.Now().Sub(scanOneDay_begin))
			}
			yearDatas.Datas[day-1] = datas[i]
			if isOld {
				fmt.Println("将会将旧文件整合到新文件:", oldFileName+" => "+fileName)
			} else if isOld2 {
				fmt.Println("将会将旧文件整合到新文件:", oldFileName2+" => "+fileName)
			}
		}

	}

	saveYearDatases(dir, loadedYears)

	return datas
	//var logCount = GetLogCount(BDUSS, forumName, time.Unix(beginUnix, 0), time.Unix(beginUnix, 0))
	//println(logCount, beginTime.Unix(), endTime.Unix())
}

func saveYearDatases(dir string, loadedYears map[string]*YearDatas) {
	for fileName, year := range loadedYears {
		func() {
			begin := time.Now()
			defer func() { fmt.Println(dir+fileName, "保存耗时", time.Now().Sub(begin)) }()

			if !year.hasChanged {
				return
			}
			f, err := os.OpenFile(dir+fileName, os.O_WRONLY, 0600)
			if os.IsNotExist(err) {
				var err error
				f, err = os.Create(dir + fileName)
				if err != nil {
					fmt.Println(dir+fileName, "无法创建文件,将不保存数据.", err)
					return
				}
			} else if err != nil {
				fmt.Println(dir+fileName, "虽然已存在,但无法打开,将不更新数据.", err)
				return
			}

			gw := gzip.NewWriter(f)
			tw := tar.NewWriter(gw)
			for _, data := range year.Datas {
				var dayJson, _ = json.Marshal(data)
				if data.Time == 0 {
					continue
				}
				header := new(tar.Header)
				header.Name = time.Unix(data.Time, 0).Format("2006-01-02.json")
				header.Size = int64(len(dayJson))

				if err := tw.WriteHeader(header); err != nil {
					fmt.Println("写入头文件失败", err)
				} else {
					_, err := tw.Write(dayJson)
					if err != nil {
						fmt.Println("写入文件失败", err)
					}
				}
			}
			tw.Flush()
			tw.Close()
			gw.Close()
			f.Close()
		}()

	}
}

func loadYearDatas(dir, fileName string, fromTime int64) *YearDatas {
	begin := time.Now()
	defer func() { fmt.Println(dir+fileName, "读取耗时", time.Now().Sub(begin)) }()
	var yearDatas *YearDatas
	if isLeap(time.Unix(fromTime, 0).Year()) {
		yearDatas = &YearDatas{make([]DayData, 366), false}
	} else {
		yearDatas = &YearDatas{make([]DayData, 365), false}
	}
	f, err := os.Open(dir + fileName)
	if err != nil {
		fmt.Println(dir+fileName, "虽然已存在,打开文件失败,进行下一步尝试.", err)
	} else {
		defer f.Close()
		gr, err := gzip.NewReader(f)
		if err != nil {
			fmt.Println(dir+fileName, "虽然已存在,打开文件失败,进行下一步尝试.", err)
		} else {
			defer gr.Close()
			tr := tar.NewReader(gr)
			for {
				header, err := tr.Next()
				if err == io.EOF {
					break
				}
				if err != nil {
					fmt.Println(dir+fileName, "中的某天虽然已存在,但是读取失败,跳过读取.", err)
				} else {
					dayName := header.FileInfo().Name()
					var year, month, day int
					_, err := fmt.Sscanf(dayName, "%d-%d-%d.json", &year, &month, &day)
					yearDay := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local).YearDay()
					if err != nil {
						fmt.Println(dir+fileName, "中某天无法解析,跳过读取.", err)
					} else {
						var dayJson, err = ioutil.ReadAll(tr)
						if err == nil {
							//fmt.Println(yearDay - 1)
							var err = json.Unmarshal(dayJson, &yearDatas.Datas[yearDay-1])
							if err != nil {
								fmt.Println(dir+fileName, "中的某天虽然已存在,但是解析json失败,跳过读取.", err)
							} else {

							}
						} else {
							fmt.Println(dir+fileName, "中的某天虽然已存在,但是读取失败,跳过读取.", err)
						}
					}

				}
			}
			return yearDatas
		}
	}
	return nil
}

func loadData(datas []DayData, i int, BDUSS, forumName string, fromTime int64, dir, fileName string, rescan bool) bool {
	if !rescan {
		f, err := os.Open(dir + fileName)
		if err != nil {
			fmt.Println(dir+fileName, "虽然已存在,打开文件失败,进行下一步尝试.", err)
		} else {
			gr, err := gzip.NewReader(f)
			if err != nil {
				fmt.Println(dir+fileName, "虽然已存在,打开文件失败,进行下一步尝试.", err)
			} else {
				tr := tar.NewReader(gr)
				_, err := tr.Next()
				if err != nil {
					fmt.Println(dir+fileName, "虽然已存在,打开文件失败,进行下一步尝试.", err)
				} else {
					var logsJson, err = ioutil.ReadAll(tr)
					if err == nil {
						var err = json.Unmarshal(logsJson, &(datas[i].Logs))
						if err == nil {
							fmt.Println(dir+fileName, "已存在,直接读取数据.")
							return true
						} else {
							fmt.Println(dir+fileName, "虽然已存在,解析Json失败,进行下一步尝试.", err)
						}
					} else {
						fmt.Println(dir+fileName, "虽然已存在,打开文件失败,进行下一步尝试.", err)
					}
				}
				gr.Close()
			}
		}
		f.Close()
	} else {
		fmt.Println(dir+fileName, "虽然已存在,但要求重新扫描,进行下一步尝试.")
		err := os.Remove(dir + fileName)
		if err != nil {
			fmt.Println(dir+fileName, "无法移除,将不删除旧数据.", err)
		}
	}
	return false
}

func loadOldData(datas []DayData, i int, BDUSS, forumName string, fromTime int64, dir, fileName string, rescan bool) bool {
	if !rescan {
		var logsJson, err = ioutil.ReadFile(dir + fileName)
		if err == nil {
			var err = json.Unmarshal(logsJson, &(datas[i].Logs))
			if err == nil {
				fmt.Println(dir+fileName, "已存在,直接读取数据.")
				return true
			} else {
				fmt.Println(dir+fileName, "虽然已存在,解析Json失败,重新扫描.", err)
			}
		} else {
			fmt.Println(dir+fileName, "虽然已存在,打开文件失败,重新扫描.", err)
		}
	} else {
		fmt.Println(dir+fileName, "虽然已存在,但要求重新扫描,重新扫描.")
		err := os.Remove(dir + fileName)
		if err != nil {
			fmt.Println(dir+fileName, "无法移除,将不删除旧数据.", err)
		}
	}
	return false
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

	var scanFinishChan = make(chan bool)

	if totalPage > 2 {
		var requireChan, doChan, finishChan = make(chan bool), make(chan bool), make(chan bool)

		go func() {
			var runningNumber int
			var requireNumber int
			var finished int
			for {
				select {
				case <-requireChan:
					if runningNumber < maxThreadNumber {
						doChan <- true
						runningNumber++
					} else {
						requireNumber++
					}
				case <-finishChan:
					finished++
					if finished == totalPage-2 {
						for i := 0; i < requireNumber; i++ {
							doChan <- false
						}
						close(requireChan)
						close(doChan)
						close(finishChan)
						scanFinishChan <- true
						return
					}
					if requireNumber > 0 {
						doChan <- true
						requireNumber--
					} else {
						runningNumber--
					}
				}
			}
		}()

		go func() {
			for i := 1; i < totalPage-1; i++ { //第一页最后一页单算
				requireChan <- true
				if <-doChan {
					go func() {
						TryGettingAndExtractLogs(BDUSS, forumName, "", OpType_None, fromTime, toTime, i, logs, 30)
						finishChan <- true
					}()

				}
			}
		}()
	}
	if totalPage >= 2 {
		TryGettingAndExtractLogs(BDUSS, forumName, "", OpType_None, fromTime, toTime, totalPage, logs, logCount-(totalPage-1)*30)
	}

	<-scanFinishChan
	return logs

}

func isLeap(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}
