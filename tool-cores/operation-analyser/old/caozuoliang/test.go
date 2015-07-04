package caozuoliang

import "github.com/purstal/pbtools/tool-cores/operation-analyser/old/log"
import "github.com/purstal/pbtools/tools/operation-analyser/csv"

//import "djimenez/iconv-go/bbbb"

//import "purstal/zhidingshanlou/core/Win8_Client"

import (
	"fmt"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

//%CE%D2%CA%C7NPC%D1%BD 我是NPC呀
//%BC%C0%D1%A9%CF%C4%D1%D7 祭雪夏炎

var local *time.Location

func init() {
	local, _ = time.LoadLocation("Local") //百度用的
}

type OPtype string

type PostLog struct {
	Author   string     //
	PostTime *time.Time //发贴时间
	TID      int        //
	PID      int        //有灵异现象,只用来记录
	PostType int        //贴子类型: 0:主题贴;1:回复与楼中楼
	Title    string     //
	Content  string     //
	OPtype   OPtype     //操作类型//ExtractData中为空
	Operator string     //操作人//ExtractData中为空
	OPtime   *time.Time //操作时间
	Duration time.Duration
}

func Do(BDUSS, _word string, bawus []Bawu, 杂项统计 []string, 整体比较 bool, 整体比较标准线 int, 整体比较白名 map[string]bool,
	begin, end *time.Time, tnf, tnc int) {

	word := UrlQueryEscape(ToGBK(_word))

	dir := "result/" + _word + "/[" + FileStringDate(begin) + "," + FileStringDate(end) + ")/"

	for _, item := range 杂项统计 {
		if eo := EnsureOperation(item); eo != 无 {
			PURSTAL(BDUSS, _word, eo, dir, begin, end, tnf, tnc)
		} else {
			log.Loglog("未找到名为", item, "的操作")
		}
	}

	bt_bawu := time.Now()
	log.Loglog("开始统计", _word, "吧删贴恢复记录")
	recover_plss := DoDo(BDUSS, word, Bawu{"", ""}, 恢复, begin, end, tnf, tnc)
	log.Loglog("完成统计", _word, "吧删贴恢复记录", "用时", time.Now().Sub(bt_bawu).String())
	WriteToCSV_fulldata(_word, StringOpType(恢复), dir, recover_plss)
	ms := ProcRecover(recover_plss)

	var sdslice []*SortedData = make([]*SortedData, len(bawus))

	for i, bawu := range bawus {
		bt_bawu := time.Now()
		log.Loglog("开始统计", _word, "吧吧务", bawu.Username)
		delete_plss := DoDo(BDUSS, word, bawu, 删贴, begin, end, tnf, tnc)
		sdslice[i] = ProcDelete(ToGBK(bawu.Username), delete_plss, begin, end, ms)
		log.Loglog("完成统计", _word, "吧吧务", bawu.Username, "用时", time.Now().Sub(bt_bawu).String())
		WriteToCSV_bawu(_word, bawu.Username, dir, begin, delete_plss, sdslice[i], i+1)
	}
	//general
	if 整体比较 {
		WriteToCSV_general(_word, dir, sdslice, 整体比较标准线, 整体比较白名)
	}

	输出整体删贴时间分布(sdslice, dir, begin)

}

func 输出整体删贴时间分布(sdslice []*SortedData, dir string, begin *time.Time) {
	f := TryCreateFile(dir+"/整体记录_删贴时间分布", "csv")
	w := csv.NewWriter(f)

	if len(sdslice) == 0 {
		return
	}

	行 := make([][]string, len(sdslice[0].Distribution)+1) //留出一行名称
	行[0] = make([]string, len(sdslice)+1)                 //留出一行显示时间
	行[0][0] = ToGBK("删贴时间分布\\吧务")

	for i, _ := range sdslice {

		行[0][i+1] = sdslice[i].Bawu_un
	}
	for i, _ := range sdslice[0].Distribution {
		行[i+1] = make([]string, len(sdslice)+1)
		t := begin.Add(time.Hour * time.Duration(i))
		行[i+1][0] = StringDate(&t)

	}
	for i, sd := range sdslice {
		for j, count := range sd.Distribution {
			行[j+1][i+1] = strconv.Itoa(count)
		}
	}
	w.WriteAll(行)

}

func NewFakeSortedData(un string) *SortedData {
	var sd SortedData
	sd.Bawu_un = un
	return &sd
}

func WriteToCSV_general(word, dir string, sdslice []*SortedData, 整体比较标准线 int, 整体比较白名 map[string]bool) {
	bt_bawu := time.Now()
	log.Loglog("开始记录", word, "吧整体记录")
	os.MkdirAll(dir, 0755) //谨慎

	//总删贴量
	//主题删贴量
	//回复删贴量
	//被恢复率(几乎都是0.00%...)

	sdslice = Remove_whitelist_sdslice(sdslice, 整体比较白名)

	sdslen := len(sdslice)

	var 主题删贴量_sum int
	var 主题删贴量 tmps = make([]tmp, sdslen+2)
	for i, sd := range sdslice {
		主题删贴量_sum += sd.ThreadCount
		主题删贴量[i] = tmp{float64(sd.ThreadCount), sd, false}
	}
	主题删贴量[sdslen] = tmp{float64(主题删贴量_sum) / float64(sdslen), NewFakeSortedData(ToGBK("平均值")), true}
	主题删贴量[sdslen+1] = tmp{主题删贴量[sdslen].i * float64(整体比较标准线) / float64(100), NewFakeSortedData(ToGBK("平均值的" + strconv.Itoa(整体比较标准线) + "%")), true}

	var 回复删贴量_sum int
	var 回复删贴量 tmps = make([]tmp, sdslen+2)
	for i, sd := range sdslice {
		回复删贴量_sum += sd.PostCount
		回复删贴量[i] = tmp{float64(sd.PostCount), sd, false}
	}
	回复删贴量[sdslen] = tmp{float64(回复删贴量_sum) / float64(sdslen), NewFakeSortedData(ToGBK("平均值")), true}
	回复删贴量[sdslen+1] = tmp{回复删贴量[sdslen].i * float64(整体比较标准线) / float64(100), NewFakeSortedData(ToGBK("平均值的" + strconv.Itoa(整体比较标准线) + "%")), true}

	var 总删贴量 tmps = make([]tmp, sdslen+2)
	for i, _ := range 主题删贴量 {
		总删贴量[i] = tmp{float64(主题删贴量[i].i + 回复删贴量[i].i), 回复删贴量[i].sd, false}
	}
	总删贴量[sdslen] = tmp{主题删贴量[sdslen].i + 回复删贴量[sdslen].i, NewFakeSortedData(ToGBK("平均值")), true}
	总删贴量[sdslen+1] = tmp{(主题删贴量[sdslen+1].i + 回复删贴量[sdslen+1].i), NewFakeSortedData(ToGBK("平均值的" + strconv.Itoa(整体比较标准线) + "%")), true}

	sort.Sort(主题删贴量)
	sort.Sort(回复删贴量)
	sort.Sort(总删贴量)

	f1 := TryCreateFile(dir+"/整体记录_删贴", "csv")
	w1 := csv.NewWriter(f1)
	//w1.Write(StringSliceToGBK([]string{"百度贴吧操作量统计工具", " by purstal ", "版本:" + Version}))

	w1.Write(StringSliceToGBK([]string{"主题删贴量:"}))
	w1.Write(StringSliceToGBK([]string{"吧务", "主题删贴量"}))
	t1 := 0
	for _, t := range 主题删贴量 {
		if t.special == true {
			w1.Write([]string{"-", t.sd.Bawu_un, strconv.FormatFloat(float64(t.i), 'f', -1, 64)}) //I wanna have UNION!!!
		} else {
			t1++
			w1.Write([]string{strconv.Itoa(t1), t.sd.Bawu_un, strconv.FormatFloat(float64(t.sd.ThreadCount), 'f', -1, 64)}) //I wanna have UNION!!!
		}

	}
	w1.Write(StringSliceToGBK([]string{}))

	w1.Write(StringSliceToGBK([]string{"回复删贴量:"}))
	w1.Write(StringSliceToGBK([]string{"吧务", "回复删贴量"}))
	t2 := 0
	for _, t := range 回复删贴量 {
		if t.special == true {
			w1.Write([]string{"-", t.sd.Bawu_un, strconv.FormatFloat(float64(t.i), 'f', -1, 64)}) //I wanna have UNION!!!
		} else {
			t2++
			w1.Write([]string{strconv.Itoa(t2), t.sd.Bawu_un, strconv.FormatFloat(float64(t.sd.PostCount), 'f', -1, 64)}) //I wanna have UNION!!!
		}

	}
	w1.Write(StringSliceToGBK([]string{}))

	w1.Write(StringSliceToGBK([]string{"总删贴量:"}))
	w1.Write(StringSliceToGBK([]string{"吧务", "总删贴量"}))
	t3 := 0
	for _, t := range 总删贴量 {
		if t.special == true {
			w1.Write([]string{"-", t.sd.Bawu_un, strconv.FormatFloat(float64(t.i), 'f', -1, 64)}) //I wanna have UNION!!!
		} else {
			t3++
			w1.Write([]string{strconv.Itoa(t3), t.sd.Bawu_un, strconv.FormatFloat(float64(t.sd.PostCount+t.sd.ThreadCount), 'f', -1, 64)}) //I wanna have UNION!!!
		}

	}

	log.Loglog("完成统计", word, "吧整体记录", "用时", time.Now().Sub(bt_bawu).String())

	w1.Flush()
	f1.Close()

	WriteToCSV(dir, "", "整体记录_删贴被恢复", func(w *csv.Writer) {

		var recover_thread_sum int
		var recover_rate_thread tmps = make([]tmp, sdslen+1)
		for i, sd := range sdslice {
			var rate float64
			if sd.ThreadCount == 0 {
				rate = 0
			} else {
				rate = float64(len(sd.RecoveredThreads)) / float64(sd.ThreadCount)
			}

			recover_thread_sum += len(sd.RecoveredThreads)
			recover_rate_thread[i] = tmp{rate, sd, false}
		}
		recover_rate_thread[sdslen] = tmp{float64(recover_thread_sum) / float64(主题删贴量_sum) / float64(sdslen), NewFakeSortedData(ToGBK("平均值")), true}

		////////
		var recover_comment_sum int
		var recover_rate_comment tmps = make([]tmp, sdslen+1)
		for i, sd := range sdslice {
			var rate float64
			if sd.PostCount == 0 {
				rate = 0
			} else {
				rate = float64(len(sd.RecoveredPosts)) / float64(sd.PostCount)
			}

			recover_comment_sum += len(sd.RecoveredPosts)
			recover_rate_comment[i] = tmp{rate, sd, false}
		}
		recover_rate_comment[sdslen] = tmp{float64(recover_comment_sum) / float64(回复删贴量_sum) / float64(sdslen), NewFakeSortedData(ToGBK("平均值")), true}
		////////
		var recover_comment_all int
		var recover_rate_all tmps = make([]tmp, sdslen+1)
		for i, t := range recover_rate_thread {
			recover_comment_all += len(t.sd.RecoveredThreads) + len(t.sd.RecoveredPosts)
			recover_rate_all[i] = tmp{(float64(len(t.sd.RecoveredThreads) + len(t.sd.RecoveredPosts))) / (float64(t.sd.ThreadCount) + float64(t.sd.PostCount)), t.sd, false}
		}
		recover_rate_all[sdslen] = tmp{float64(recover_comment_sum+recover_thread_sum) / float64(回复删贴量_sum+主题删贴量_sum), NewFakeSortedData(ToGBK("平均值")), true}

		sort.Sort(recover_rate_thread)
		sort.Sort(recover_rate_comment)
		sort.Sort(recover_rate_all)

		w.Write(StringSliceToGBK([]string{"主题删贴被恢复率:"}))
		w.Write(StringSliceToGBK([]string{"", "吧务", "主题删贴被恢复率", "主题删贴被恢复量", "主题删贴量"}))
		t4 := 0
		for _, t := range recover_rate_thread {
			if t.special == true {
				w.Write([]string{"-", t.sd.Bawu_un, strconv.FormatFloat(float64(t.i)*100, 'f', -1, 64) + "%", strconv.Itoa(recover_thread_sum), strconv.Itoa(回复删贴量_sum)}) //I wanna have UNION!!!
			} else {
				t4++
				w.Write([]string{strconv.Itoa(t4), t.sd.Bawu_un, strconv.FormatFloat(float64(t.i)*100, 'f', -1, 64) + "%", strconv.Itoa(len(t.sd.RecoveredThreads)), strconv.Itoa(t.sd.ThreadCount)}) //I wanna have UNION!!!
			}

		}
		w.Write(StringSliceToGBK([]string{}))

		w.Write(StringSliceToGBK([]string{"回复删贴被恢复率:"}))
		w.Write(StringSliceToGBK([]string{"", "吧务", "回复删贴被恢复率", "回复删贴被恢复量", "回复删贴量"}))
		t5 := 0
		for _, t := range recover_rate_comment {
			if t.special == true {
				w.Write([]string{"-", t.sd.Bawu_un, strconv.FormatFloat(float64(t.i)*100, 'f', -1, 64) + "%", strconv.Itoa(recover_comment_sum), strconv.Itoa(主题删贴量_sum)}) //I wanna have UNION!!!
			} else {
				t5++
				w.Write([]string{strconv.Itoa(t5), t.sd.Bawu_un, strconv.FormatFloat(float64(t.i)*100, 'f', -1, 64) + "%", strconv.Itoa(len(t.sd.RecoveredPosts)), strconv.Itoa(t.sd.PostCount)}) //I wanna have UNION!!!
			}

		}
		w.Write(StringSliceToGBK([]string{}))

		w.Write(StringSliceToGBK([]string{"总删贴被恢复率:"}))
		w.Write(StringSliceToGBK([]string{"", "吧务", "总删贴被恢复率", "总删贴被恢复量", "总删贴量"}))
		t6 := 0
		for _, t := range recover_rate_all {
			if t.special == true {
				w.Write([]string{"-", t.sd.Bawu_un, strconv.FormatFloat(float64(t.i)*100, 'f', -1, 64) + "%", strconv.Itoa(recover_comment_sum + recover_thread_sum), strconv.Itoa(回复删贴量_sum + 主题删贴量_sum)}) //I wanna have UNION!!!
			} else {
				t6++
				w.Write([]string{strconv.Itoa(t6), t.sd.Bawu_un, strconv.FormatFloat(float64(t.i)*100, 'f', -1, 64) + "%", strconv.Itoa(len(t.sd.RecoveredThreads) + len(t.sd.RecoveredPosts)), strconv.Itoa(t.sd.PostCount + t.sd.ThreadCount)}) //I wanna have UNION!!!
			}

		}
	})

}

type tmp struct {
	i       float64
	sd      *SortedData
	special bool
}

type tmps []tmp

func (t tmps) Len() int {
	return len(t)
}
func (t tmps) Less(i, j int) bool {
	return t[i].i > t[j].i
}
func (t tmps) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func PURSTAL(BDUSS, word string, op_type OPtype, dir string, begin, end *time.Time, tnf, tnc int) {
	bt_bawu := time.Now()

	stroptype := StringOpType(op_type)

	log.Loglog("开始统计", word, "吧", stroptype, "记录")

	plss := DoDo(BDUSS, word, Bawu{"", ""}, op_type, begin, end, tnf, tnc)
	log.Loglog("完成统计", word, "吧", stroptype, "记录", "用时", time.Now().Sub(bt_bawu).String())
	WriteToCSV_fulldata(word, stroptype, dir, plss)

}

func WriteToCSV_fulldata(word, item_name, dir string, plss [][]*PostLog) {
	bt_bawu := time.Now()
	log.Loglog("开始记录", word, "吧", item_name, "记录")

	os.MkdirAll(dir, 0755)
	f1 := TryCreateFile(dir+"/完整"+item_name+"记录", "csv")
	//f1.WriteString("\xEF\xBB\xBF")
	w1 := csv.NewWriter(f1)

	w1.Write(StringSliceToGBK([]string{"百度贴吧操作量统计工具", " by purstal ", "版本:" + Version}))
	w1.Write(StringSliceToGBK([]string{"发贴人", "可能的发贴时间", "tid", "pid", "贴子类型", "标题", "预览", "操作人", "操作日期"}))
	for _, pls := range plss {
		for _, pl := range pls {
			w1.Write([]string{pl.Author, StringDate(pl.PostTime), strconv.Itoa(pl.TID), EnsurePID(pl.PID), StringPostType(pl.PostType), pl.Title, pl.Content, pl.Operator, StringDate(pl.OPtime)})
		}

	}
	w1.Flush()
	f1.Close()

	log.Loglog("完成记录", word, "吧", item_name, "记录", "用时", time.Now().Sub(bt_bawu).String())

}

func ProcRecover(plss [][]*PostLog) []map[int]*PostLog {
	ms := make([]map[int]*PostLog, 2)
	ms[0] = make(map[int]*PostLog)
	ms[1] = make(map[int]*PostLog)
	for _, pls := range plss {
		for _, pl := range pls {
			if pl.PostType == 主题 {
				ms[pl.PostType][pl.TID] = pl
			} else {
				ms[pl.PostType][pl.PID] = pl
			}
		}
	}
	return ms
}

func DoDo(BDUSS, word string, bawu Bawu, op_type OPtype, begin, end *time.Time, tnf, tnc int) [][]*PostLog { //tn:threadsnum
	resp, _ := ListPostLog_op(BDUSS, word, bawu.UrlEncoding, op_type, begin, end, 1)
	//rx := regexp.MustCompile(`tbui_pagination_right">\D*(\d*)`)
	//rx := regexp.MustCompile(`tbui_total_page">\D*(\d*)`)
	rx := regexp.MustCompile(`<div class="breadcrumbs">.*?<em>(\d*)</em>.*?</div>`)
	var logCount int
	if len(rx.FindStringSubmatch(resp)) == 2 {
		logCount, _ = strconv.Atoi(rx.FindStringSubmatch(resp)[1])
	}
	if logCount == 0 {
		return make([][]*PostLog, 0)
	}

	lastPageLogCount := logCount - logCount/30*30
	var pagecount int
	if lastPageLogCount != 0 {
		pagecount = logCount/30 + 1
	} else {
		pagecount = logCount / 30
	}

	a, b := Allocate(pagecount-1, tnf)

	var from int = 1
	var to int = 0

	finish := make(chan bool, tnf-1)
	defer close(finish)

	if a == 0 {
		tnf = b
	}
	var plss [][]*PostLog = make([][]*PostLog, tnf+1)

	procfunc := func(i, from, to, logCountShouldHave int) {
		ch := make(chan *PostLog, 30*tnc-1)
		finish2 := make(chan bool, tnc-1)
		defer close(finish2)
		go LoadData(BDUSS, word, bawu, op_type, begin, end, from, to, ch, logCountShouldHave)
		var tmp [][]*PostLog = make([][]*PostLog, tnc)
		for j := 0; j < tnc; j++ {
			//意味不明
			go func(j int) {
				tmp[j] = ProcessData(ch)
				finish2 <- true
			}(j)

		}
		for j := 0; j < tnc; {
			<-finish2
			j++
		}
		plss[i] = JoinPLS(tmp)
		//if !isLastPage {
		finish <- true
		//}

	}

	for i := 0; i < tnf; i++ {
		if i < b {
			to = from + a + 1
		} else {
			to = from + a
		}
		if from == to {
			continue
		}
		go procfunc(i, from, to, 30)
		from = to
	}
	if lastPageLogCount != 0 {
		procfunc(tnf, pagecount, pagecount+1, lastPageLogCount)
		for i := 0; i < tnf+1; {
			<-finish
			i++
			fmt.Println("进度:", i, "/", tnf+1)
		}
	} else {
		for i := 0; i < tnf; {
			<-finish
			i++
			fmt.Println("进度:", i, "/", tnf)
		}
	}

	return plss

}

func LoadData(BDUSS, word string, bawu Bawu, op_type OPtype, begin, end *time.Time, from, to int, ch chan *PostLog, logCountShouldHave int) {

	rx2 := regexp.MustCompile(`"post_author"><.*?>(.*?)<.*?">(.*?)<.*?href="/p/(.*?)\?pid=(.*?)#.*?title="(�ظ���)?(.*?)".*?"post_text">            (.*?)<.*?".*?">(.*?)<.*?"ui_text_normal">(.*?)</a></td><td>(.*?)<br/>(.*?)<`)
	//stringslice[1]发贴人
	//stringslice[2]发贴时间
	//stringslice[3]tid
	//stringslice[4]pid
	//stringslice[5]贴子类型
	//stringslice[6]标题
	//stringslice[7]内容
	//stringslice[8]操作人
	//stringslice[9]操作类型
	//stringslice[10]操作日期
	//stringslice[11]操作时间

	for i := from; i < to; i++ {
		for retryTimes := 0; ; retryTimes++ {
			j := 0
			resp, err := ListPostLog_op(BDUSS, word, bawu.UrlEncoding, op_type, begin, end, i)
			if err != nil {
				//logorz(TMP, err)
				var pl PostLog
				pl.Author = "ERROR"
				ch <- &pl
				continue
			}

			stringsliceslice := rx2.FindAllStringSubmatch(resp, -1)
			if (len(stringsliceslice) != logCountShouldHave) || len(stringsliceslice) == 0 {
				if retryTimes < 0 {
					log.Loglog("第", j+1, "次获取日志第", i, "页数量异常, 实际", len(stringsliceslice), "条. 无重试次数上限.") //("第", j+1, "次获取到空日志,无重试次数上限.页数:", i)
					continue
				} /*else if retryTimes == 空日志重试次数 {
					log.Loglog("第", j+1, "次获取日志第", i, "页数量异常, 实际", len(stringsliceslice), "条. 超过重试次数上限.")
					break
				} else {
					log.Loglog("第", j+1, "次获取日志第", i, "页数量异常, 实际", len(stringsliceslice), "条. 次数上限为", 空日志重试次数, "次.")
					continue
				}*/
			}

			for _, stringslice := range stringsliceslice {
				tid, _ := strconv.Atoi(stringslice[3])
				pid, _ := strconv.Atoi(stringslice[4])
				pyear := EnsureYear(tid)
				var pmonth, pday, phour, pmin int
				fmt.Sscanf(stringslice[2], "%d��%d�� %d:%d", &pmonth, &pday, &phour, &pmin)

				posttime := time.Date(pyear, time.Month(pmonth), pday, phour, pmin, 0, 0, local)
				//operation := EnsureOperation(stringslice[5])
				var oyear, omonth, oday, ohour, omin int
				var posttype int
				if stringslice[5] != "" {
					posttype = 1
				} else {
					posttype = 0
				}
				fmt.Sscanf(stringslice[10], "%d-%d-%d", &oyear, &omonth, &oday)
				fmt.Sscanf(stringslice[11], "%d:%d", &ohour, &omin)

				optime := time.Date(oyear, time.Month(omonth), oday, ohour, omin, 0, 0, local)

				if posttype == 主题 {
					pid = -1
				}
				pl := &PostLog{stringslice[1], &posttime, tid, pid, posttype, stringslice[6], stringslice[7], EnsureOperation(stringslice[8]), stringslice[9], &optime, optime.Sub(posttime)}

				ch <- pl
				//fmt.Println("ExtractData", pl)
			}
			break
		}

	}
	close(ch)

}

func Allocate(pagecount, threadsnum int) (int, int) {
	//a:少的
	//b:多的的数量

	a := pagecount / threadsnum
	b := pagecount - (a * threadsnum)

	return a, b

}

func ProcDelete(un string, plss [][]*PostLog, begin, end *time.Time, ms []map[int]*PostLog) *SortedData {
	data := Iamunknown(un, plss, int(math.Ceil(end.Sub(*begin).Hours())), begin, ms)
	return SortData(data)
}

func Iamunknown(un string, plss [][]*PostLog, hour int, begin *time.Time, ms []map[int]*PostLog) *Data {

	var data = NewData(hour)
	data.Bawu_un = un
	for _, pls := range plss {
		for _, pl := range pls {
			if _, found := data.SameThreads[pl.TID]; !found {
				data.SameThreads[pl.TID] = new(SameThread)
				data.SameThreads[pl.TID].ThreadTitle = pl.Title
			}
			if pl.PostType == 主题 {
				data.ThreadCount++
				data.SameThreads[pl.TID].ThreadIsDeletedByTheOperator = true
				if ms[主题][pl.TID] != nil {
					data.RecoveredThreads = append(data.RecoveredThreads, Recovered{ms[主题][pl.TID].OPtime, ms[主题][pl.TID].Operator, pl})
				}

			} else {
				data.PostCount++
				data.SameThreads[pl.TID].Count++
				if ms[回复][pl.PID] != nil {
					data.RecoveredPosts = append(data.RecoveredPosts, Recovered{ms[回复][pl.PID].OPtime, ms[回复][pl.PID].Operator, pl})
				}
			}

			if pl.Duration > 旧贴间隔 {
				data.OldPosts = append(data.OldPosts, pl)
				data.Speed[len(SpeedClass)]++
			} else {
				for i, class := range SpeedClass {
					if pl.Duration < class {
						data.Speed[i]++
						break
					}
				}

			}

			data.SameAccounts[pl.Author] = append(data.SameAccounts[pl.Author], pl)

			data.Distribution[int(pl.OPtime.Sub(*begin).Hours())]++

		}
	}
	return data

}

func JoinPLS(plss [][]*PostLog) (pls golang) {
	for _, x := range plss {
		pls = append(pls, x...)
	}
	sort.Sort(pls)
	return []*PostLog(pls)

}

type golang []*PostLog

func (g golang) Len() int {
	return len(g)
}

func (g golang) Less(i, j int) bool {
	return g[i].OPtime.After(*g[j].OPtime)
}

func (g golang) Swap(i, j int) {
	g[i], g[j] = g[j], g[i]
}

func FileStringDate(t *time.Time) string {
	return strconv.Itoa(t.Year()) + "-" + strconv.Itoa(int(t.Month())) + "-" + strconv.Itoa(t.Day())
}

func TryCreateFile(name, extension string) *os.File {
	var err1, err2 error
	var f *os.File
	f, err1 = os.Create(name + "." + extension)
	if err1 != nil {
		log.Loglog(name+"."+extension,
			"创建失败:\n", err1.Error(), "\n尝试创建为",
			name+strconv.FormatInt(time.Now().Unix(), 10)+"."+extension)
		f, err2 = os.Create(name + strconv.FormatInt(time.Now().Unix(), 10) + "." + extension)
		if err2 != nil {
			log.Loglog(name+"."+extension,
				"创建失败:\n", err1.Error(), "\n跳过写入此文件")
		}
	}
	return f
}

func StringDate_Day(t *time.Time) string {
	return strconv.Itoa(t.Year()) + "-" + strconv.Itoa(int(t.Month())) + "-" + strconv.Itoa(t.Day())

}

func MyMkdir(dir, foldername string) string {
	if foldername != "" {
		dir = dir + foldername + "/"
	}

	os.Mkdir(dir, 0755)
	return dir
}

func WriteToCSV(dir, foldername, filename string, wfunc func(*csv.Writer)) {
	newdir := MyMkdir(dir, foldername)
	f := TryCreateFile(newdir+filename, "csv")
	w := csv.NewWriter(f)
	w.Write(StringSliceToGBK([]string{"百度贴吧操作量统计工具", " by purstal ", "版本:" + Version}))
	wfunc(w)
	w.Flush()
	f.Close()

}

func WriteToCSV_bawu(word, bawu, dir string, begin *time.Time, plss [][]*PostLog, sd *SortedData, num int) {

	bt_bawu := time.Now()
	log.Loglog("开始记录", word, "吧吧务", bawu)

	filename := fmt.Sprintf("%02d.", num) + bawu

	//f1.WriteString("\xEF\xBB\xBF")

	/*
		WriteToCSV(dir, "删贴速度", filename, func(w *csv.Writer) {

		})
	*/
	WriteToCSV(dir, "##完整删贴记录##", filename+"_完整删贴记录", func(w *csv.Writer) {
		w.Write(StringSliceToGBK([]string{"发贴人", "可能的发贴时间", "tid", "pid", "贴子类型", "标题", "预览", "操作日期"}))
		for _, pls := range plss {
			for _, pl := range pls {
				w.Write([]string{pl.Author, StringDate(pl.PostTime), strconv.Itoa(pl.TID), EnsurePID(pl.PID), StringPostType(pl.PostType), pl.Title, pl.Content, StringDate(pl.OPtime)})
			}

		}
	})

	//////
	allcount := sd.ThreadCount + sd.PostCount
	WriteToCSV(dir, "删贴速度", filename+"_删贴速度", func(w *csv.Writer) {
		w.Write(StringSliceToGBK([]string{"吧务用户名:", bawu}))
		w.Write(StringSliceToGBK([]string{"删贴总量:", strconv.Itoa(allcount), "主题贴数量:", strconv.Itoa(sd.ThreadCount), "回复贴数量:", strconv.Itoa(sd.PostCount), "主题贴占比:", strconv.FormatFloat((float64(sd.ThreadCount)/float64(sd.ThreadCount+sd.PostCount)*100), 'f', 2, 64) + "%"}))

		w.Write([]string{})
		w.Write(StringSliceToGBK([]string{"删贴速度:", "(最大误差1分钟)"}))

		w.Write([]string{"[0," + ToGBK(SpeedClassString[0]) + ")", strconv.Itoa(sd.Speed[0])})

		for i := 1; i < len(SpeedClass); i++ {
			w.Write(StringSliceToGBK([]string{"[" + SpeedClassString[i-1] + "," + SpeedClassString[i] + ")", strconv.Itoa(sd.Speed[i])}))

		}
		w.Write([]string{"[" + ToGBK(SpeedClassString[len(SpeedClassString)-1]) + ",+" + ToGBK("∞") + ")", strconv.Itoa(sd.Speed[len(SpeedClass)])})

	})

	////////

	h_slice := make([]int, 24)
	WriteToCSV(dir, "删贴时间分布_小时", filename+"_删贴时间分布_小时", func(w *csv.Writer) {

		w.Write(StringSliceToGBK([]string{"删贴时间分布:", "(以小时为单位)"}))
		for i, v := range sd.Distribution {
			t := begin.Add(time.Hour * time.Duration(i))
			w.Write([]string{StringDate(&t), strconv.Itoa(v)})
		}

	})

	////////

	WriteToCSV(dir, "删贴时间分布_日", filename+"_删贴时间分布_日", func(w *csv.Writer) {
		w.Write(StringSliceToGBK([]string{"删贴时间分布:", "(以日为单位)"}))
		for i := 0; i < len(sd.Distribution); i += 24 {
			var sum int
			for h, v := range sd.Distribution[i : i+24] {
				sum += v
				h_slice[h] += v
			}

			t := begin.Add(time.Hour * time.Duration(i))
			w.Write([]string{StringDate_Day(&t), strconv.Itoa(sum)})
		}
	})

	////////

	WriteToCSV(dir, "同主题内删贴", filename+"_同主题内删贴", func(w *csv.Writer) {
		j := 0
		ss := [][]string{}
		if 同主题判定 != 0 {
			for _, st := range sd.SameThreads {
				if st.Count >= 同主题判定 {
					ss = append(ss, []string{st.ThreadTitle, strconv.Itoa(st.Count), BoolToString(st.ThreadIsDeletedByTheOperator), strconv.FormatFloat((float64(st.Count)/float64(allcount)*100), 'f', 2, 64) + "%"})
				}
			}
		}

		w.Write(StringSliceToGBK([]string{"相同主题内删贴:", strconv.Itoa(j), "个主题内"}))
		w.Write(StringSliceToGBK([]string{"标题", "次数", "之后本人删掉主题?", "占总删贴量比"}))
		w.WriteAll(ss)
	})

	////////

	WriteToCSV(dir, "删贴被恢复", filename+"_删贴被恢复", func(w *csv.Writer) {
		w.Write(StringSliceToGBK([]string{"被恢复的主题:", strconv.Itoa(len(sd.RecoveredThreads)), "占总主题删贴量比", strconv.FormatFloat((float64(len(sd.RecoveredPosts))/float64(sd.ThreadCount)*100), 'f', -1, 64) + "%",
			"被恢复的回复删贴:", strconv.Itoa(len(sd.RecoveredPosts)), "占总回复删贴量比", strconv.FormatFloat((float64(len(sd.RecoveredPosts)*100)/float64(sd.PostCount*100)), 'f', -1, 64) + "%"}))
		w.Write(StringSliceToGBK([]string{"发贴人", "可能的发贴时间", "删贴日期", "贴子类型", "标题", "预览", "恢复人", "最后恢复日期"}))
		for _, r := range sd.RecoveredThreads {
			w.Write([]string{r.pl.Author, StringDate(r.pl.PostTime), StringDate(r.pl.OPtime), ToGBK("主题"), r.pl.Title, r.pl.Content, r.recoverby, StringDate(r.recovertime)})
		}
		for _, r := range sd.RecoveredPosts {
			w.Write([]string{r.pl.Author, StringDate(r.pl.PostTime), StringDate(r.pl.OPtime), ToGBK("回复"), r.pl.Title, r.pl.Content, r.recoverby, StringDate(r.recovertime)})
		}
	})

	////////

	WriteToCSV(dir, "疑似删除的旧贴", filename+"_疑似删除的旧贴", func(w *csv.Writer) {
		w.Write(StringSliceToGBK([]string{"疑似删除的旧贴:", "(发贴到删贴间隔超过" + 旧贴间隔.String() + "):", strconv.Itoa(len(sd.OldPosts))}))
		w.Write(StringSliceToGBK([]string{"发贴人", "可能的发贴时间", "操作时间", "时间间隔", "贴子类型", "标题", "预览"}))
		for _, op := range sd.OldPosts {
			w.Write([]string{op.Author, StringDate(op.PostTime), StringDate(op.OPtime), op.Duration.String(), StringPostType(op.PostType), op.Title, op.Content})
		}
	})

	////////

	WriteToCSV(dir, "相同用户", filename+"_相同用户", func(w *csv.Writer) {
		i := 0
		ii := 0
		ss := [][]string{}
		if 同账号判定 != 0 {
			for _, pls := range sd.SameAccounts {
				if len(pls) >= 同账号判定 {
					i++
					ii += len(pls)
					ss = append(ss, []string{pls[0].Author, strconv.Itoa(len(pls)), strconv.FormatFloat((float64(len(pls))/float64(allcount)*100), 'f', 2, 64) + "%"})
				}
			}
		}
		w.Write(StringSliceToGBK([]string{"相同用户:", strconv.Itoa(i), "个用户 共计", strconv.Itoa(ii), "个贴子"}))
		w.Write(StringSliceToGBK([]string{"用户名", "次数", "占总删贴量比"}))
		w.WriteAll(ss)
	})

	log.Loglog("完成记录", word, "吧吧务", bawu, "用时", time.Now().Sub(bt_bawu).String())

}

var 是_GBK = ToGBK("是")
var 否_GBK = ToGBK("否")

func BoolToString(b bool) string {
	if b == true {
		return 是_GBK
	} else {
		return 否_GBK
	}
}

func StringDate(t *time.Time) string {
	return strconv.Itoa(t.Year()) + "-" + strconv.Itoa(int(t.Month())) + "-" + strconv.Itoa(t.Day()) + " " + strconv.Itoa(t.Hour()) + ":" + strconv.Itoa(t.Minute())
}

func EnsurePID(pid int) string {
	if pid == -1 {
		return ""
	}
	return strconv.Itoa(pid)
}

func GetBawuList(BDUSS, _word string, whitelist map[string]bool) (bawuslice []Bawu) {
	word := UrlQueryEscape(ToGBK(_word))

	t2 := time.Now()
	t1 := time.Date(1970, 1, 1, 0, 0, 0, 0, local)
	resp, _ := ListPostLog_op(BDUSS, word, "", 删贴, &t1, &t2, 1)

	rx := regexp.MustCompile(`"no-pointer">(.*?)<`)
	stringsliceslice := rx.FindAllStringSubmatch(resp, -1)

	if len(stringsliceslice) == 0 {
		//Loglog("获取吧务名单失败")
		return bawuslice
	}

	for _, Bawu_un := range stringsliceslice[1:] {
		if whitelist[strings.ToLower(FromGBK(Bawu_un[1]))] != true {
			bawu := Bawu{UrlQueryEscape(Bawu_un[1]), FromGBK(Bawu_un[1])}
			bawuslice = append(bawuslice, bawu)
		}
	}
	return bawuslice
}

type Bawu struct {
	UrlEncoding string
	Username    string
}

func GetBawuList_C(slice []string) (bawuslice []Bawu) {
	for _, str := range slice {
		bawuslice = append(bawuslice, Bawu{UrlQueryEscape(ToGBK(str)), str})
	}
	return bawuslice
}

func Remove_whitelist_sdslice(sdslice []*SortedData, whitelist map[string]bool) (_sdslice []*SortedData) {
	for _, sd := range sdslice {
		if whitelist[strings.ToLower(FromGBK(sd.Bawu_un))] != true {
			_sdslice = append(_sdslice, sd)
		}

	}
	return _sdslice
}
