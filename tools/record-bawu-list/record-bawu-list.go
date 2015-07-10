package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/PuerkitoBio/goquery"

	"github.com/purstal/pbtools/modules/misc"
)

const URLTemplet = "http://tieba.baidu.com/bawu2/platform/listBawuTeamInfo?word=%s"

type BawuInfo struct {
	Title   string
	Members []string
}

type Record struct {
	BawuInfos []BawuInfo
}

func main() {
	if len(os.Args) < 2 {
		panic("wrong usage.")
	}

	var forum = os.Args[1]

	var URL = fmt.Sprintf(URLTemplet, misc.UrlQueryEscape(misc.ToGBK(forum)))

	var doc *goquery.Document

	for {
		_doc, err := goquery.NewDocument(URL)
		if err == nil {
			doc = _doc
			break
		}
	}

	var record Record

	doc.Find(`div.bawu_single_type`).Each(func(i int, sel *goquery.Selection) {
		var info BawuInfo
		info.Title = misc.FromGBK(sel.Find(`div.title`).Text())
		sel.Find(`span.member`).Each(func(i int, member *goquery.Selection) {
			info.Members = append(info.Members, misc.FromGBK(member.Find(`a.user_name`).Text()))
		})
		record.BawuInfos = append(record.BawuInfos, info)
	})

	var data, _ = json.Marshal(record)
	fmt.Print(string(data))

}
