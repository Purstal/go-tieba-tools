package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	//"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/purstal/pbtools/modules/logs"
	"github.com/purstal/pbtools/modules/misc"
)

type Record struct {
	Title   string `json:"标题"`
	IsReply bool   `json:"回复"`
	Content string `json:"摘要"`
	Url     string `json:"url"`
	Forum   string `json:"贴吧"`
	Time    string `json:"时间"`
}

func main() {
	if len(os.Args) != 2 {
		panic("wrong usage")
	}
	userName := os.Args[1]

	var postCount = -1

	const rn = 100

	var records []Record

	var retryTimes = 0
	var flag = false
	//var lastLen = -1
	for pn := 0; (postCount == -1 || postCount >= pn+rn) && !flag; {
		var doc *goquery.Document

		logs.Info("获取贴子中,pn:", pn)

		for {
			var resp, err = rGetUPost(userName, pn, rn)
			if err == nil {
				body, err := ioutil.ReadAll(resp.Body)
				if err == nil {
					body = ([]byte)(misc.FromGBK(string(body)))
					if err == nil {
						doc, err = goquery.NewDocumentFromReader(bytes.NewReader(body))
						if err == nil {
							break
						}
					}
				}
			}
		}

		trs := doc.Find(`div#tabarea`).Find(`tr`)
		trsLen := trs.Length()
		logs.Info("获取到贴数:", trsLen-1)
		if trsLen != rn+1 {
			if retryTimes < 10 {
				logs.Info("重试,重试次数:", retryTimes)
				retryTimes++
				continue
			} else {
				if trsLen > 90 {
					logs.Info("啊..也许是百度抽了也说不定.")
				} else {
					flag = true
					logs.Info("重试达到上限,本次后结束.")
				}

			}
		}
		trs.Each(func(i int, sel *goquery.Selection) {
			if i == 0 {
				if pn == 0 {
					postCountStr := sel.Find(`td.padL10`).Find(`font`).Text() //有问题,但是,我不管了!
					postCountStr = strings.TrimPrefix(postCountStr, "(共")
					postCountStr = strings.TrimSuffix(postCountStr, "条)")
					_postCount, err := strconv.Atoi(postCountStr)
					if err == nil && (_postCount != 0 || trsLen == 0) {
						logs.Info("贴子数:", postCount)
						logs.Info("预计页数:", (postCount - 1/100))
						postCount = postCount
					} else {
						if err != nil {
							logs.Info("无法获取贴子数:", err)
						} else {
							logs.Info("无法获取贴子数,postCount:", _postCount)
						}
					}
				} else {
					return
				}
			}

			var record Record

			list := sel.Find(`td.list`)
			list_a := list.Find(`a`)
			title := list_a.Text()
			if strings.HasPrefix(title, "回复：") {
				record.IsReply = true
				title = strings.TrimPrefix(title, "回复：")
			}
			record.Title = title
			record.Url, _ = list_a.Attr(`href`)
			record.Content = list.Find(`div`).Text()

			listtd_1 := sel.Find(`td.listtd`)
			record.Forum = listtd_1.First().Text()

			record.Time = listtd_1.Next().Text()

			records = append(records, record)
		})

		retryTimes = 0
		pn += 100

	}

	f, _ := os.Create(userName + ".result.json")

	json.NewEncoder(f).Encode(struct {
		TotalPostCount       int      `json:"总计贴数"`
		RealFetchedPostCount int      `json:"实际获取贴数"`
		Records              []Record `json:"记录"`
	}{postCount, len(records), records})
}

func rGetUPost(userName string, pn, rn int) (*http.Response, error) {
	return http.Get("http://tieba.baidu.com/f/upost" +
		"?un=" + misc.UrlQueryEscape(misc.ToGBK(userName)) +
		"&pn=" + strconv.Itoa(pn) +
		"&rn=" + strconv.Itoa(rn))
}
