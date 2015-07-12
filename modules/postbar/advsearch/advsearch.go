package advsearch

import (
	"fmt"
	//"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/purstal/pbtools/modules/http"
	"github.com/purstal/pbtools/modules/logs"
	"github.com/purstal/pbtools/modules/misc"
	"github.com/purstal/pbtools/modules/postbar"
)

func GetAdvSearchResultPage(kw, un string, rn int, pn int) (string, error) {
	//120字符,不算上"回复："
	var parameters http.Parameters

	//parameters.Add("ie", "utf-8")
	parameters.Add("ie", "gbk")

	parameters.Add("kw", misc.ToGBK(kw))
	parameters.Add("rn", strconv.Itoa(rn))
	parameters.Add("un", misc.ToGBK(un))
	parameters.Add("sm", "1")
	if pn != 1 {
		parameters.Add("pn", strconv.Itoa(pn))
	}

	resp, err := http.Get(`http://tieba.baidu.com/f/search/ures`, parameters, nil)
	if err != nil {
		return "", err
	}

	return string(resp), nil
}

type AdvSearchResult struct {
	postbar.IPost
	Tid      uint64
	Pid      uint64
	IsReply  bool
	Title    string
	Content  string
	Forum    string
	PostTime time.Time
	Author   AdvSearchAuthor
}

type AdvSearchAuthor struct {
	postbar.IAuthor
	Name string
}

type AdvSearchThread struct {
	postbar.IThread
	Tid   uint64
	Title string
}

func (a AdvSearchAuthor) AGetID() (bool, uint64)                         { return false, 0 }
func (a AdvSearchAuthor) AGetName() string                               { return a.Name }
func (a AdvSearchAuthor) AGetIsLike() (bool, bool)                       { return false, false }
func (a AdvSearchAuthor) AGetLevel() (bool, uint8)                       { return false, 0 }
func (a AdvSearchAuthor) AGetPortrait() (bool, string)                   { return false, "" }
func (a AdvSearchResult) PGetPid() uint64                                { return a.Pid }
func (a AdvSearchResult) PGetFloor() (bool, int)                         { return false, 0 }
func (a AdvSearchResult) PGetPostTime() (bool, time.Time)                { return false, a.PostTime }
func (a AdvSearchResult) PGetOriginalContentList() (bool, []interface{}) { return false, nil }
func (a AdvSearchResult) PGetContentList() []postbar.Content {
	return []postbar.Content{postbar.Text{a.Content}}
}
func (a AdvSearchResult) PContentIsComplete() bool    { return false } //如果只有文字,<=120
func (a AdvSearchResult) PGetAuthor() postbar.IAuthor { return a.Author }

func (t AdvSearchThread) TGetTid() uint64                      { return t.Tid }
func (t AdvSearchThread) TGetTitle() string                    { return t.Title }
func (t AdvSearchThread) TGetReplyNum() (bool, uint32)         { return false, 0 }
func (t AdvSearchThread) TGetLastReplyTime() (bool, time.Time) { return false, time.Time{} }
func (t AdvSearchThread) TGetIsTop() (bool, bool)              { return false, false }
func (t AdvSearchThread) TGetIsGood() (bool, bool)             { return false, false }
func (t AdvSearchThread) TGetAuthor() postbar.IAuthor          { return nil }
func (t AdvSearchThread) TGetLastReplyer() postbar.IAuthor     { return nil }
func (t AdvSearchThread) TGetContentList() []postbar.Content {
	return nil
}
func (t AdvSearchThread) TContentIsComplete() bool { return false }

func GetAdvSearchResultList(kw, un string, rn,
	pn int) ([]AdvSearchResult, error) {
	resp, err := GetAdvSearchResultPage(kw, un, rn, pn)

	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp))
	if err != nil {
		return nil, err
	}

	return ParseAdvSearchDocument(doc), nil
}

func ParseAdvSearchDocument(doc *goquery.Document) []AdvSearchResult {
	posts := doc.Find(`div.s_post_list`).Eq(0).Find(`div.s_post`)
	results := make([]AdvSearchResult, posts.Length())
	posts.Each(func(index int, post *goquery.Selection) {
		result := &results[index]
		bluelink := post.Find(`a.bluelink`)
		title := bluelink.Text()
		oldLen := len(title)
		result.Title = strings.TrimPrefix(misc.FromGBK(title), `回复：`)
		result.IsReply = oldLen != len(result.Title)
		link, _ := bluelink.Attr(`href`)

		ids := idreg.FindStringSubmatch(link)

		if len(ids) != 3 {
			logs.Error("高级搜索结果链接异常,跳过:", link, ".")
			goto CONTINUE
		}
		{
			result.Tid, _ = strconv.ParseUint(ids[1], 10, 64)
			result.Pid, _ = strconv.ParseUint(ids[2], 10, 64)

			result.Content = misc.FromGBK(post.Find(`div.p_content`).Text())

			date := misc.FromGBK(post.Find(`font.p_date`).Text())
			var y, m, d, hour, min int
			fmt.Sscanf(date, "%d-%d-%d %d:%d", &y, &m, &d, &hour, &min)
			result.PostTime = time.Date(y, time.Month(m), d, hour, min, 0, 0, time.Local)

			x := post.Find(`font.p_violet`)
			result.Forum = misc.FromGBK(x.First().Text())
			result.Author.Name = misc.FromGBK(x.Next().Text())

		}

	CONTINUE:
	})

	return results
}

var advreg = regexp.MustCompile(`<span class="p_title"><a class="bluelink" href="/p/(.*?)\?pid=(.*?)&.*?>(回复：)?(.*?)</a>.*?"p_content">(.*?)</div>.*?p_date">(.*?)<`)
var idreg = regexp.MustCompile(`/p/(\d+).*?pid=(\d+)`)

//没有Forum
func OldGetAdvSearchResultList(kw, un string, rn,
	pn int) ([]AdvSearchResult, error) {
	resp, err := GetAdvSearchResultPage(kw, un, rn, pn)

	if err != nil {
		return nil, err
	}
	results := advreg.FindAllStringSubmatch(misc.FromGBK(resp), -1)
	if len(results) == 0 {
		return nil, nil
	}

	asrs := make([]AdvSearchResult, len(results))

	for i, result := range results {
		asr := &asrs[i]
		asr.Tid, _ = strconv.ParseUint(result[1], 10, 64)
		asr.Pid, _ = strconv.ParseUint(result[2], 10, 64)
		asr.IsReply = result[3] == `回复：`
		asr.Title = result[4]
		asr.Content = result[5]

		var y, m, d, hour, min int
		fmt.Sscanf(result[6], "%d-%d-%d %d:%d", &y, &m, &d, &hour, &min)
		asr.PostTime = time.Date(y, time.Month(m), d, hour, min, 0, 0, time.Local)

		asr.Author.Name = un

	}

	return asrs, nil

}

//<span class="p_title"><a class="bluelink" href="/p/(.*?)\?pid=(.*?)&.*?>(回复：)?(.*?)</a>.*?"p_content">(.*?)</div>.*?p_date">(.*?)<
/*
http://tieba.baidu.com/f/search/ures?ie=utf-8&kw=x&qw=&rn=10&un=y&sm=1
//仅用于无关键词!
*/
