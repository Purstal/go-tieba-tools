package scanner

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"
)

func saveMonthesDatas(dir string, loadedMonths map[string]*MonthDatas) {
	for fileName, month := range loadedMonths {
		saveOneMonthDatas(dir, fileName, month)
	}
}

func saveOneMonthDatas(dir string, fileName string, month *MonthDatas) {
	begin := time.Now()
	if !month.hasChanged {
		return
	}

	defer func() { fmt.Println(dir+fileName, "有修改,保存.保存耗时", time.Now().Sub(begin)) }()
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
	for _, data := range month.Datas {
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
}

func loadMonthDatas(dir, fileName string, fromTime int64) *MonthDatas {
	begin := time.Now()
	defer func() { fmt.Println(dir+fileName, "读取耗时", time.Now().Sub(begin)) }()
	var monthDatas *MonthDatas

	var goFromTime = time.Unix(fromTime, 0)
	if m := int(goFromTime.Month()); m != 2 || isLeap(goFromTime.Year()) {
		monthDatas = &MonthDatas{make([]DayData, COMMON_YEAR_MONTH_DAYS[m]), false}
	} else {
		monthDatas = &MonthDatas{make([]DayData, 29), false}
	}

	f, err := os.Open(dir + fileName)
	if err != nil {
		fmt.Println(dir+fileName, "虽然已存在,打开文件失败,重新扫描.", err)
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
					//yearDay := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local).YearDay()
					if err != nil {
						fmt.Println(dir+fileName, "中某天无法解析,跳过读取.", err)
					} else {
						var dayJson, err = ioutil.ReadAll(tr)
						if err == nil {
							var err = json.Unmarshal(dayJson, &monthDatas.Datas[day-1])
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
			return monthDatas
		}
	}
	return nil
}
