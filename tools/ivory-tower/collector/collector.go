package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/purstal/pbtools/modules/logs"
	"github.com/purstal/pbtools/modules/postbar"
	"github.com/purstal/pbtools/modules/postbar/apis/forum-win8-1.5.0.0"
	"github.com/purstal/pbtools/modules/postbar/apis/thread-win8-1.5.0.0"
	"github.com/purstal/pbtools/tools/ivory-tower/collector/collects"
	"github.com/purstal/pbtools/tools/ivory-tower/collector/records/save-to-html"
	"github.com/purstal/pbtools/tools/operation-analyser/csv"
)

func main() {

	var begin = time.Now()

	os.Mkdir("log", 0644)
	f, err0 := os.Create("log/" + begin.Format("2006-01-02 15-04-05.log"))
	if err0 != nil {
		logs.Error("无法创建log文件:", err0)
	} else {
		logs.SetDefaultLogger(logs.NewLogger(logs.DebugLevel, os.Stdout, f))
	}

	flags := parseFlag()
	if flags == nil {
		printUsage()
		return
	}

	if flags.FromTime == nil {
		flags.FromTime = &begin
	}

	var accWin8 = postbar.NewDefaultWindows8Account("")

	logs.Info("收集贴子.")
	var records = Collect(accWin8, flags.ForumName, flags.ToTime).ToReversedSortedSlice()
	logs.Info("收集完毕.")

	{
		logs.Info("验证是否新主题.")
		var n = len(records)
		for i, record := range records {
			logs.Info(fmt.Sprintf("第%d个主题,最多%d个主题.主题tid:%d.", i, n, record.Tid))
			var postTime = GetPostTime(TryGettingThread(accWin8, record.Tid, 2, 1))
			if postTime == nil {
				continue
			}
			records[i].PostTime = postTime.Unix()
			if postTime.Before(flags.ToTime) || record.Tid <= flags.MinTid {
				records = records[:i]
				break
			}
		}
		logs.Info("验证完毕.")
	}

	logs.Info("记录数据.")

	os.Mkdir("新主题", 0644)
	if flags.OutputFormat.CSV {
		if err := SaveRecords("./新主题/", records, begin, flags.ToTime); err != nil {
			logs.Info("csv记录失败:", err)
		} else {
			logs.Info("csv记录完毕.")
		}
	}

	if flags.OutputFormat.HTML {
		var xRecords = make([]save_to_html.Record, len(records))
		for i, record := range records {
			xRecords[i] = save_to_html.Record{
				Tid:      record.Tid,
				Title:    record.Title,
				Author:   record.Author,
				Time:     time.Unix(record.PostTime, 0).Format("2006-01-02 15:04:05"),
				Abstract: save_to_html.ExtractAbstract(record.Abstract),
			}
		}
		const TIME_LAYOUT = "2006年01月02日15点04分05秒"
		const DATE_LAYOUT = "2006年01月02日"
		if flags.Merge {
			var timeStr = fmt.Sprintf("%s 至 %s", begin.Format(TIME_LAYOUT), flags.ToTime.Format(TIME_LAYOUT))
			f, err := os.Create("./新主题/" + timeStr + ".html")
			if err != nil {
				logs.Info("html记录失败:", err)
			} else {
				save_to_html.Write(f, xRecords, timeStr)
				logs.Info("html记录完毕.")
				f.Close()
			}
		} else {
			var i int
			var n = len(records)
			for day := flags.ToTime.Unix(); i < n; day += 24 * 60 * 60 {
				var boundary = day + 24*60*60
				var f *os.File
				var err error
				var timeStr string
				if boundary < begin.Unix() {
					timeStr = fmt.Sprintf("%s", time.Unix(day, 0).Format(DATE_LAYOUT))
					f, err = os.Create("./新主题/" + timeStr + ".html")
				} else {
					timeStr = fmt.Sprintf("%s 截至 %s", time.Unix(day, 0).Format(DATE_LAYOUT), begin.Format("15点04分05秒"))
					f, err = os.Create("./新主题/" + timeStr + ".html")
				}
				var j int
				var breakFlag bool
				for ; j < n && records[j].PostTime < boundary; j++ {
					if records[j].PostTime > flags.ToTime.Unix() {
						breakFlag = true
					}
				}
				if err != nil {
					logs.Info("html记录失败:", err)
				} else {
					save_to_html.Write(f, xRecords[i:j], timeStr)
					logs.Info("html记录完毕.")
					f.Close()
				}
				if breakFlag {
					break
				}
			}
		}

	}

	logs.Info("消耗时间:", time.Now().Sub(begin))

}

type Flags struct {
	ForumName string
	ToTime    time.Time  //必选
	FromTime  *time.Time //可选
	MinTid    uint64
	Merge     bool

	OutputFormat struct {
		CSV, HTML, JSON bool
	}
}

func parseFlag() *Flags {
	var flags = Flags{
		OutputFormat: struct {
			CSV, HTML, JSON bool
		}{CSV: true, HTML: true},
	}
	if len(os.Args) == 2 {
		return nil
	}
	flags.ForumName = os.Args[1]
	logs.Info("贴吧:", os.Args[1])
	var toTime, err1 = parseTime(os.Args[2])
	if err1 != nil {
		logs.Fatal(fmt.Sprintf("截止日期格式有误: %s.", err1.Error()))
		return nil
	}
	flags.ToTime = toTime

	logs.Info("截止日期:", os.Args[2])

	tryParsingBoolFlag := func(b *bool, name string, arg string) bool {
		if arg == "--"+name {
			*b = true
			return true
		} else if strings.HasPrefix(arg, "--"+name+"=") {
			v := strings.TrimPrefix(arg, "--"+name+"=")
			r, err := strconv.ParseBool(v)
			if err == nil {
				*b = r
			}
			return true
		}
		return false
	}

	for _, arg := range os.Args[3:] {
		if strings.HasPrefix(arg, "--from-time=") {
			v := strings.TrimPrefix(arg, "--from-time=")
			if fromTime, err := parseTime(v); err != nil {
				logs.Fatal(fmt.Sprintf("起始日期格式有误: %s.", err1.Error()))
				return nil
			} else {
				flags.FromTime = &fromTime
				logs.Info("设置起始日期:", fromTime.Format("2006-01-02 15:04:05"))
			}
		} else if strings.HasPrefix(arg, "--min-tid=") {
			v := strings.TrimPrefix(arg, "--min-tid=")
			var minTid, err = strconv.ParseUint(v, 10, 64)
			if err != nil {
				logs.Fatal(fmt.Sprintf("最小tid格式有误: %s.", err.Error()))
				return nil
			}
			flags.MinTid = minTid
			logs.Info("设置最小tid:", minTid)
		} else if tryParsingBoolFlag(&flags.Merge, "merge", arg) {
		} else if tryParsingBoolFlag(&flags.OutputFormat.CSV, "CSV", arg) {
		} else if tryParsingBoolFlag(&flags.OutputFormat.HTML, "HTML", arg) {
		} else if tryParsingBoolFlag(&flags.OutputFormat.JSON, "JSON", arg) {
		} else {
			logs.Warn("未知flag:", arg)
		}
	}

	return &flags
}

type Config struct {
	RN                  int `json:"每页扫描请求贴数"`
	CollectThreadNumber int `json:"收集线程数"`
	ComfirmThreadNumber int `json:"验证线程数"`
}

func printUsage() {
	fmt.Sprintf(`usage:
%s 截止日期 [最小tid(不记录此tid)]
例: %s minecraft 2006-01-02
例: %s minecraft 2006-01-02 1234567890
`, os.Args[0], os.Args[0], os.Args[0])
}

func Collect(accWin8 *postbar.Account, forumName string, to time.Time) ThreadMap {
	var set = ThreadMap{} //见下面
	collect(accWin8, forumName, to, set)
	return set
}

func collect(accWin8 *postbar.Account, forumName string, to time.Time, set ThreadMap) {
	logs.Debug("本轮收集至:", to.String())
	const RN = 100
	var lastTime time.Time
	for pn := 1; ; pn++ {
		logs.Debug("本轮收集pn:", pn)
		var threads = TryGettingForumPageThreads(accWin8, forumName, RN, pn)
		if len(threads) != 0 {
			logs.Debug(fmt.Sprintf("本页第一贴时间%s; 剩余:%s", threads[0].LastReplyTime.Format("2006-01-02 15:04:05"), threads[0].LastReplyTime.Sub(to).String()))
		} else {
			logs.Warn(fmt.Sprintf("经多次尝试,本页一贴没有,返回. pn:%d", pn))
			return
		}
		for i, theThread := range threads {
			if theThread.LastReplyTime.After(lastTime) {
				lastTime = theThread.LastReplyTime
			}
			if theThread.LastReplyTime.Before(to) {
				if pn == 1 && i < len(threads)-1 {
					return
				}
				collect(accWin8, forumName, lastTime, set) //保证毫无缺漏
				return
			}
			if _, exist := set[theThread.Tid]; !exist {
				set[theThread.Tid] = collects.ThreadRecord{
					Title:    theThread.Title,
					Tid:      theThread.Tid,
					Author:   theThread.Author.Name,
					Abstract: append(theThread.Abstract, theThread.MediaList...), /*extractAbstract(append(theThread.Abstract, theThread.MediaList...))*/
				}
			}
		}
	}
	logs.Debug("结束本轮收集.")
}

func GetPostTime(theThread *thread.Thread) *time.Time {
	if theThread == nil {
		return nil
	}
	return &theThread.PostList[0].PostTime
}

func TryGettingThread(accWin8 *postbar.Account, tid uint64, rn, pn int) *thread.Thread {
	for {
		t, err, pberr := thread.GetThread2(accWin8, tid, false, 0, pn, rn, false, false, false)
		if err == nil {
			if pberr == nil || pberr.ErrorCode == 0 {
				if len(t.PostList) != 0 {
					return t
				}
				logs.Error(fmt.Sprintf("无法获取主题回复,重试."), pn, pberr.ErrorCode, pberr.ErrorMsg)
			} else if pberr.ErrorCode == 4 {
				logs.Error(fmt.Sprintf("主题已被删除,跳过. tid:%d; pn:%d; pberror:%d(%s).", tid, pn, pberr.ErrorCode, pberr.ErrorMsg))
				return nil
			} else {
				logs.Error(fmt.Sprintf("无法获取主题,重试. tid:%d; pn:%d; pberror:%d(%s).", tid, pn, pberr.ErrorCode, pberr.ErrorMsg))
			}
		} else {
			logs.Error(fmt.Sprintf("无法获取主题,重试. tid:%d; pn:%d; error:%s", tid, pn, err.Error()))
		}
	}
}

func TryGettingForumPageThreads(accWin8 *postbar.Account, forumName string, rn, pn int) []*forum.ForumPageThread {
	var retryTimes int
	const MAX_RETRY_TIME = 2
	for {
		_, posts, _, err, pberr := forum.GetForumStruct(accWin8, forumName, rn, pn)
		if err == nil {
			if pberr == nil || pberr.ErrorCode == 0 {
				if len(posts) != rn {
					if rn < len(posts) {
						logs.Warn(fmt.Sprintf("本页所求rn:%d, 实际rn:%d", rn, len(posts)))
						return posts
					}
					if retryTimes < MAX_RETRY_TIME {
						retryTimes++
					} else {
						logs.Warn(fmt.Sprintf("经%d次尝试, 本页所求rn:%d, 实际rn:%d", MAX_RETRY_TIME+1, rn, len(posts)))
						return posts
					}
				} else {
					return posts
				}
			} else {
				logs.Error(fmt.Sprintf("无法获取主页,重试. pn:%d; pberror:%d(%s).", pn, pberr.ErrorCode, pberr.ErrorMsg))
			}
		} else {
			logs.Error(fmt.Sprintf("无法获取主页,重试. pn:%d; error:%s", pn, err.Error()))
		}
	}
}

type ThreadMap map[uint64]collects.ThreadRecord

func (set ThreadMap) ToReversedSortedSlice() []collects.ThreadRecord {
	var slice = make([]collects.ThreadRecord, len(set))
	var i = 0
	for _, v := range set {
		slice[i] = v
		i++
	}
	sort.Sort(sort.Reverse(ThreadSliceSorter(slice)))
	return slice
}

type ThreadSliceSorter []collects.ThreadRecord

func (s ThreadSliceSorter) Less(i, j int) bool {
	return s[i].Tid < s[j].Tid
}

func (s ThreadSliceSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ThreadSliceSorter) Len() int {
	return len(s)
}

func parseTime(str string) (time.Time, error) { //time.Parse的location不对
	var y, m, d int
	_, err := fmt.Sscanf(str, "%d-%d-%d", &y, &m, &d)
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.Local), err
}

//从 `../recorder/recorder.go` 复制来的
func extractAbstract(contents []interface{}) string {
	if len(contents) == 0 {
		return "摘要: []"
	}
	var str = "摘要: ["

	for i, _content := range contents {
		content, ok1 := _content.(map[string]interface{})
		t, ok2 := content["type"].(string)
		if !ok1 {
			t = "解析失败"
		} else if !ok2 {
			t = "缺少类型"
		}

		switch t {
		case "0": //文字
			if content["text"] != nil {
				str += fmt.Sprintf("#文字: %s", fmt.Sprint(content["text"]))
			} else {
				str += fmt.Sprintf("#文字: 解析失败(%s)", fmt.Sprint(content))
			}
		case "3": //图片
			if content["big_pic"] != nil {
				str += fmt.Sprintf("#图片: %s", fmt.Sprint(content["big_pic"]))
			} else {
				str += fmt.Sprintf("#图片: 解析失败(%s)", fmt.Sprint(content))
			}
		case "5": //视频
			if content["vhsrc"] != nil {
				str += fmt.Sprintf("#视频: %s", fmt.Sprint(content["vhsrc"]))
				if content["vsrc"] != nil {
					str += fmt.Sprint("(%s)", content["vsrc"])
				}
			} else if content["vsrc"] != nil {
				str += fmt.Sprintf("#视频: %s", fmt.Sprint(content["vsrc"]))
			} else {
				str += fmt.Sprintf("#视频: 解析失败(%s)", fmt.Sprint(content))
			}
		case "6":
			if content["src"] != nil {
				str += fmt.Sprintf("#音乐: %s", fmt.Sprint(content["src"]))
			} else {
				str += fmt.Sprintf("#音乐: 解析失败(%s)", fmt.Sprint(content))
			}
		case "解析失败":
			str += fmt.Sprintf("#未知: 解析失败(%s)", fmt.Sprint(content))
		case "缺少类型":
			str += fmt.Sprintf("#未知: 缺少类型(%s)", fmt.Sprint(content))
		default:
			str += fmt.Sprintf("#未知: 未知类型(%s)(%s)", t, fmt.Sprint(content))
		}
		if i != len(contents)-1 {
			str += "; "
		} else {
			str += "]"
		}
	}

	return str
}

//从 `../recorder/recorder.go` 复制来并稍作修改
func SaveRecords(path string, records []collects.ThreadRecord, from, to time.Time) error {
	const TIME_FORMAT_LAYOUT = "2006年01月02日15点04分05秒"
	f, err := os.Create(path + fmt.Sprintf("%s 至 %s.csv", from.Format(TIME_FORMAT_LAYOUT), to.Format(TIME_FORMAT_LAYOUT)))
	f.Write([]byte{0xEF, 0xBB, 0xBF})
	if err != nil {
		return err
	}
	w := csv.NewWriter(f)

	w.Write([]string{fmt.Sprintf("收集时间: %s => %s", from, to)})
	w.Write(nil)

	for _, r := range records {
		w.WriteAll([][]string{[]string{fmt.Sprintf("tid: %d", r.Tid), fmt.Sprintf("作者: %s", r.Author),
			fmt.Sprintf("标题: %s", r.Title)}, []string{fmt.Sprintf("时间: %s", time.Unix(r.PostTime, 0).Format("2006-01-02 15:04:05")), extractAbstract(r.Abstract)}, nil})
	}
	w.Flush()
	f.Close()
	return nil
}
