package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"

	"github.com/PuerkitoBio/goquery"

	"github.com/purstal/go-tieba-base/tieba/adv-search"
	"github.com/purstal/pbtools/tools/operation-analyser/csv"
)

type Record struct {
	TID      uint64
	Title    string
	IsThread bool
	Pid      uint64
	Forum    string
	Time     string
}

func main() {
	const un = `iamunknown`
	const upper_pid = 22214552705
	//http://tieba.baidu.com/p/1730481997?pid=22214552705&cid=#22214552705

	var rest []Record

OUTER:
	for pn := 1; ; pn++ {
		fmt.Println(pn)
		var results []advsearch.AdvSearchResult
		var retryTimes int
		for {
			url := fmt.Sprintf(`http://tieba.baidu.com/f/search/ures?kw=&qw=&sm=0&rn=10&un=%s&pn=%d`, un, pn)
			doc, _ := goquery.NewDocument(url)
			results = advsearch.ParseAdvSearchDocument(doc)
			if len(results) == 10 || retryTimes == 10 {
				retryTimes = 0
				break
			} else {
				retryTimes++
			}
		}

		for _, result := range results {
			if result.Pid == upper_pid {
				break OUTER
			}
			if !result.IsReply {
				fmt.Println(10000)
			}
			rest = append(rest, Record{result.Tid, result.Title, !result.IsReply, result.Pid, result.Forum, result.PostTime.Format("2006-01-02 15:04")})
		}
	}

	fmt.Println(len(rest))

	records := append(rest, importOldResults()...)

	f_posts, _ := os.Create("iamunknown.posts.csv")
	f_threads, _ := os.Create("iamunknown.thread.csv")
	f_json, _ := os.Create("iamunknown.json")

	f_posts.Write([]byte{0xEF, 0xBB, 0xBF})
	f_threads.Write([]byte{0xEF, 0xBB, 0xBF})

	w_posts := csv.NewWriter(f_posts)
	w_threads := csv.NewWriter(f_threads)

	w_posts.Write([]string{"tid", "标题", "贴吧", "pid", "时间"})
	w_threads.Write([]string{"tid", "标题", "贴吧", "pid", "时间"})

	for _, record := range records {
		if record.IsThread {
			w_threads.Write([]string{strconv.FormatUint(record.TID, 10), record.Title, record.Forum, strconv.FormatUint(record.Pid, 10), record.Time})
		} else {
			w_posts.Write([]string{strconv.FormatUint(record.TID, 10), record.Title, record.Forum, strconv.FormatUint(record.Pid, 10), record.Time})
		}
	}
	w_posts.Flush()
	f_posts.Close()
	w_threads.Flush()
	f_threads.Close()

	data, _ := json.Marshal(records)
	f_json.Write(data)

}

var exp = regexp.MustCompile(`/p/(\d*)\?pid=(\d*)`)
var exp2 = regexp.MustCompile(`(.*)\d\d\d\d-\d\d-\d\d`)

func importOldResults() []Record {
	data, _ := ioutil.ReadFile(`iamunknown.result.json`)

	var oldRecords struct {
		Records []struct {
			Title   string `json:"标题"`
			IsReply bool   `json:"回复"`
			Content string `json:"摘要"`
			Url     string `json:"url"`
			Forum   string `json:"贴吧"`
			Time    string `json:"时间"`
		} `json:"记录"`
	}

	json.Unmarshal(data, &oldRecords)

	var records = make([]Record, 0, len(oldRecords.Records))

	for i := len(oldRecords.Records) - 1; i >= 0; i-- {
		oldRecord := &oldRecords.Records[i]
		var record Record

		matchResult := exp.FindStringSubmatch(oldRecord.Url)

		record.TID, _ = strconv.ParseUint(matchResult[1], 10, 64)
		record.Pid, _ = strconv.ParseUint(matchResult[2], 10, 64)
		record.Title = oldRecord.Title
		record.IsThread = !oldRecord.IsReply
		record.Time = oldRecord.Time

		matchResult2 := exp2.FindStringSubmatch(oldRecord.Forum)

		record.Forum = matchResult2[1]

		records = append(records, record)
	}

	return records

}
