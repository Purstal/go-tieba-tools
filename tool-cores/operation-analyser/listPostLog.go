package scaner

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	//"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/purstal/go-tieba-base/http"
	"github.com/purstal/go-tieba-base/misc"
	//"github.com/purstal/pbtools/tool-cores/operation-analyser/old/log"
)

type OpType int

const (
	OpType_None       OpType = 0
	OpType_Delete     OpType = 12
	OpType_Recover    OpType = 13
	OpType_AddGood    OpType = 17
	OpType_CancelGood OpType = 18
	OpType_AddTop     OpType = 25
	OpType_CancelTop  OpType = 26
)

func ListPostLog(BDUSS, forumName, svalue string, opType OpType, fromTime, toTime int64, pn int) ([]byte, error) {

	var parameters http.Parameters

	parameters.Add("word", misc.UrlQueryEscape(misc.ToGBK(forumName)))
	parameters.Add("op_type", strconv.Itoa(int(opType)))
	parameters.Add("stype", "op_uname") //post_uname:发贴人;op_uname:操作人
	parameters.Add("svalue", svalue)    //utf8-url-encoding
	parameters.Add("date_type", "on")   //限定时间范围
	if fromTime != 0 {
		parameters.Add("begin", strconv.FormatInt(fromTime, 10)) //起始时间戳
	}
	if toTime != 0 {
		parameters.Add("end", strconv.FormatInt(toTime, 10)) //结束时间戳
	}
	parameters.Add("pn", strconv.Itoa(pn))

	var cookies http.Cookies

	cookies.Add("BDUSS", BDUSS)

	return http.Get("http://tieba.baidu.com/bawu2/platform/listPostLog", parameters, cookies)
}

func TryListingPostLog(BDUSS, forumName, svalue string, opType OpType, fromTime, toTime int64, pn int) []byte {
	for {
		var resp, err = ListPostLog(BDUSS, forumName, svalue, opType, fromTime, toTime, pn)
		if err == nil {
			return []byte(misc.FromGBK(string(resp)))
		}
	}
}

func TryGettingListingPostLogDocument(BDUSS, forumName, svalue string, opType OpType, fromTime, toTime int64, pn int) *goquery.Document {
	for {
		var doc, err = goquery.NewDocumentFromReader(bytes.NewReader(TryListingPostLog(BDUSS, forumName, svalue, opType, fromTime, toTime, pn)))
		if err == nil {
			return doc
		}
	}
}

func CanViewBackstage(doc *goquery.Document) bool {
	return doc.Find(`div#operator_menu`).Length() >= 1
}

func ExtractLogCount(doc *goquery.Document) (int, error) {
	var count, err = strconv.Atoi(doc.Find(`div.breadcrumbs`).Contents().Filter(`em`).Text())

	return count, err
}

type Log struct {
	Author   string
	PostTime struct {
		Month, Day, Hour, Minute int
	}
	Title       string
	IsReply     bool
	Text        string
	MediaHtml   string
	TID         int
	PID         int
	OperateType OpType
	Operator    string
	OperateTime struct {
		Year, Month, Day, Hour, Minute int
	}
}

var extractFromLinkRegexp *regexp.Regexp

func init() {
	extractFromLinkRegexp = regexp.MustCompile(`/p/(\d*)\?pid=(\d*)`)
}

func ExtractLogs(doc *goquery.Document, logs []Log, fromIndex, toIndex int) (trueLength bool) {
	var logTrs = doc.Find(`table.data_table`).Find(`tbody`).Find(`tr`)

	if logTrs.Length() != toIndex-fromIndex {
		return false
	}

	logTrs.Each(func(i int, tr *goquery.Selection) {
		//var log = &logs[fromIndex+i]
		var log = &logs[fromIndex+i] //
		if len(logs) <= fromIndex+i {
			panic(fmt.Sprintln(len(logs), fromIndex+i, fromIndex, toIndex))
		}
		log.Author = tr.Find(`div.post_author`).Find(`a`).Text()
		fmt.Sscanf(tr.Find(`time.ui_text_desc`).Text(),
			"%d月%d日 %d:%d",
			&log.PostTime.Month, &log.PostTime.Day,
			&log.PostTime.Hour, &log.PostTime.Minute)
		var contentSel = tr.Find(`div.post_content`)
		var a = contentSel.Find(`h1`).Find(`a`)
		var title = strings.TrimSpace(a.Text())
		if strings.HasPrefix(title, "回复：") {
			log.IsReply = true
			log.Title = strings.TrimPrefix(title, "回复：")
		} else {
			log.Title = title
		}
		var herf, _ = a.Attr(`href`)
		fmt.Sscanf(herf, "/p/%d?pid=%d", &log.TID, &log.PID)

		log.Text = strings.TrimSpace(contentSel.Find(`div.post_text`).Text())
		var mediaHtml, _ = contentSel.Find(`div.post_media`).Html()
		if strings.TrimSpace(log.MediaHtml) != "" {
			log.MediaHtml = mediaHtml
		}

		var td2 = tr.Find(`td`).Next()
		switch td2.Find(`span`).Text() {
		case "删贴":
			log.OperateType = OpType_Delete
		case "恢复":
			log.OperateType = OpType_Recover
		case "加精":
			log.OperateType = OpType_AddGood
		case "取消加精":
			log.OperateType = OpType_CancelGood
		case "置顶":
			log.OperateType = OpType_AddTop
		case "取消置顶":
			log.OperateType = OpType_CancelTop
		default:
			log.OperateType = OpType_None
		}

		var td3 = td2.Next()
		log.Operator = td3.Find(`a.ui_text_normal`).Text()

		var td4 = td3.Next()
		var td4Html, _ = td4.Html()
		fmt.Sscanf(td4Html,
			`%d-%d-%d<br/>%d:%d`,
			&log.OperateTime.Year,
			&log.OperateTime.Month, &log.OperateTime.Day,
			&log.OperateTime.Hour, &log.OperateTime.Minute)
	})

	return true

}

func TryGettingAndExtractLogs(BDUSS, forumName, svalue string,
	opType OpType, fromTime, toTime int64,
	pn int, logs []Log, trueCount int) {
	for {
		var doc = TryGettingListingPostLogDocument(BDUSS, forumName, svalue, opType, fromTime, toTime, pn)
		if ExtractLogs(doc, logs, (pn-1)*30, (pn-1)*30+trueCount) {
			return
		}
	}

}
