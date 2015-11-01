package save_to_html

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/purstal/go-tieba-tools/ivory-tower/collector/collects"
)

type PageData struct {
	PageID      string
	TimeRange   string
	ThreadLists []ThreadList
}

type ThreadList struct {
	Date       string
	ThreadList []Thread
}

type Thread struct {
	Title     string
	Tid       uint64
	Author    string
	Abstracts []Abstract
}

type Abstract struct {
	Type    string
	Content string
}

func SaveRecods(dir string, threads []collects.Thread, rangeStr string) {
	var pageData PageData
	pageData.PageID = strconv.FormatInt(time.Now().UnixNano())
	pageData.TimeRange = rangeStr
	if len(threads) == 0 {
		pageData.ThreadLists = []ThreadList{}
	} else {
		pageData.ThreadLists = []ThreadList{
			ThreadList{Date: time.Unix(threads[0].GetTime(), 0).Format("2006-01-02"),
				ThreadList: make([]Thread, len(threads))},
		}
	}

}
