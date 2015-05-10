package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	//"path"
	"strconv"
	"strings"
	"time"

	"github.com/purstal/pbtools/modules/postbar"
	"github.com/purstal/pbtools/modules/postbar/thread-win8-1.5.0.0"
	"github.com/purstal/pbtools/tools/operation-analyser/csv"
)

type ThreadX struct {
	Version string
	Info    struct {
		Author string
		Title  string
	}
	ScanInfo struct {
		FromTime time.Time
		ToTime   time.Time
		ScanTime time.Time
	}
	PostList []thread.ThreadPagePost
}

func main() {
	var usage = fmt.Sprintf(`usage:
%s $filename

$filename: `, os.Args[0])

	if len(os.Args) != 2 {
		fmt.Println(usage)
		return
	}

	var data []byte

	{
		var err error
		if data, err = ioutil.ReadFile(os.Args[1]); err != nil {
			fmt.Println("无法读取文件:", err, ".")
		}
	}

	var tx ThreadX

	{
		err := json.Unmarshal(data, &tx)
		if err != nil {
			fmt.Println("无法解析json:", err, ".")
		}
	}

	analyse(tx)

}

type UpdateRecord struct {
	Time time.Time

	Times     int
	TextCount int
	PicCount  int
}

func analyse(tx ThreadX) {
	if len(tx.PostList) == 0 {
		fmt.Println("这期间根本没贴子,分析个毛线,退出.")
		return
	}

	//ft := tx.PostList[0].PostTime

	var hourRecord []*UpdateRecord
	var dayRecord []*UpdateRecord
	var weekRecord []*UpdateRecord
	var monthRecord []*UpdateRecord

	var t time.Time

	for _, post := range tx.PostList {
		_t := t
		t = post.PostTime
		if !isSameHour(_t, t) {
			hourRecord = append(hourRecord, &UpdateRecord{Time: time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, time.Local)})
			if !isSameDay(_t, t) {
				dayRecord = append(dayRecord, &UpdateRecord{Time: time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)})
				if !isSameWeek(_t, t) {
					dayBegin := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
					weekDay := int(dayBegin.Weekday())
					weekBeginDay := dayBegin.Truncate(time.Duration((weekDay-1)*24) * time.Hour)
					weekRecord = append(weekRecord, &UpdateRecord{Time: weekBeginDay})
				}
				if !isSameMonth(_t, t) {
					monthRecord = append(monthRecord, &UpdateRecord{Time: time.Date(t.Year(), t.Month(), 0, 0, 0, 0, 0, time.Local)})
				}
			}
		}

		var textCount, picCount int

		for _, content := range post.PGetContentList() {
			switch content.(type) {
			case (postbar.Text):
				textCount += len(content.(postbar.Text).Text)
			case (postbar.Pic):
				picCount++
			}

		}
		record(textCount, picCount, hourRecord, dayRecord, weekRecord, monthRecord)
	}

	table := makeTable(hourRecord, "小时记录")
	table = append(table, makeTable(dayRecord, "日记录")...)
	table = append(table, makeTable(weekRecord, "周记录")...)
	table = append(table, makeTable(monthRecord, "月记录")...)

	var dir = "analyse-result/"

	if err := os.MkdirAll(dir, 0644); err != nil {
		fmt.Println("创建目录失败,将不保存结果:", err, ".")
		return
	}

	slice := strings.Split(os.Args[1], "/")

	f, err1 := os.Create(dir + slice[len(slice)-1] + ".csv")
	if err1 != nil {
		fmt.Println("创建文件失败,将不保存结果:", err1, ".")
		return
	}
	w := csv.NewWriter(f)
	err2 := w.WriteAll(table)
	if err2 != nil {
		fmt.Println("保存csv失败,将不保存结果:", err2, ".")
		return
	}
	w.Flush()
	return
}

func record(textCount, picCount int, collectors ...[]*UpdateRecord) {
	for _, collector := range collectors {
		r := collector[len(collector)-1]
		r.Times++
		r.TextCount += textCount
		r.PicCount += picCount
	}
}

func makeTable(records []*UpdateRecord, recordType string) [][]string {
	var table = make([][]string, len(records)+2)
	table[0] = []string{recordType + ":", "更新次数", "更新文字数", "更新图片数", "更新图文比"}
	for i := 0; i < len(table)-2; i++ {
		var columnA string
		switch recordType {
		case "小时记录":
			columnA = records[i].Time.Format("2006年01月02日 15时")
		case "日记录":
			columnA = records[i].Time.Format("2006年01月02日")
		case "周记录":
			columnA = records[i].Time.Format("2006年01月02周")
		case "月记录":
			columnA = records[i].Time.Format("2006年01月")
		}
		var rate string
		if records[i].TextCount != 0 {
			rate = strconv.FormatFloat(float64(records[i].PicCount)/float64(records[i].TextCount), 'f', 4, 64)
		}
		table[i+1] = []string{columnA, strconv.Itoa(records[i].Times),
			strconv.Itoa(records[i].TextCount), strconv.Itoa(records[i].PicCount), rate}

	}
	return table
}

func isSameDay(a, b time.Time) bool {
	d := a.Sub(b).Hours()
	if d < 0 {
		d = -d
	}
	if d < 24 && a.Day() == b.Day() {
		return true
	}
	return false
}

func isSameHour(a, b time.Time) bool {
	d := a.Sub(b).Minutes()
	if d < 0 {
		d = -d
	}
	if d < 60 && a.Hour() == b.Hour() {
		return true
	}
	return false
}

func isSameMonth(a, b time.Time) bool {
	d := a.Sub(b).Hours()
	if d < 0 {
		d = -d
	}
	if d < 100*24 && a.Month() == b.Month() {
		return true
	}
	return false
}

func isSameWeek(a, b time.Time) bool {
	dayBeginA := time.Date(a.Year(), a.Month(), a.Day(), 0, 0, 0, 0, time.Local)
	weekDayA := int(dayBeginA.Weekday())
	weekBeginDayA := dayBeginA.Truncate(time.Duration((weekDayA-1)*24) * time.Hour)
	dayBeginB := time.Date(b.Year(), b.Month(), b.Day(), 0, 0, 0, 0, time.Local)
	weekDayB := int(dayBeginB.Weekday())
	weekBeginDayB := dayBeginB.Truncate(time.Duration((weekDayB-1)*24) * time.Hour)

	if weekBeginDayA.Equal(weekBeginDayB) {
		return true
	}
	return false
}
