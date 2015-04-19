package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	_ "net/http/pprof"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/purstal/pbtools/modules/logs"
	"github.com/purstal/pbtools/modules/postbar"
	"github.com/purstal/pbtools/modules/postbar/accounts"
	"github.com/purstal/pbtools/modules/postbar/advsearch"
	"github.com/purstal/pbtools/modules/postbar/apis"
	postfinder "github.com/purstal/pbtools/tools-core/post-finder"
	"purstal/pbtools2/tools/pbutil"
)

var accWin8 *accounts.Account
var accAndr *accounts.Account

func init() {
	go func() {
		http.ListenAndServe(":33101", nil)
	}()
}

func LoadSettings(fileName string) (*Settings, error) {

	file, err := os.Open(fileName)

	if err != nil {
		return nil, err
	}

	data, err2 := ioutil.ReadAll(file)
	if err2 != nil {
		return nil, err2
	}
	if data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		data = data[3:]
	}

	var settings Settings

	err3 := json.Unmarshal(data, &settings)
	if err3 != nil {
		return nil, err3
	}
	return &settings, nil
}

type Settings struct {
	BDUSS                     string
	ForumName                 string `json:"贴吧"`
	PostContentRegexpFilePath string `json:"贴子内容正则文件"`
	//DefaultWaterThreadTids    []uint64 `json:"默认水楼tids"`
	BawuList []string `json:"吧务列表"`
}

var settings *Settings

func main() {
	os.Mkdir("logs", 0777)
	logfile, _ := os.Create("logs/" + strconv.FormatInt(time.Now().Unix(), 16))
	defer logfile.Close()
	logs.DefaultLogger = logs.NewLogger(logs.DebugLevel, os.Stdout, logfile)
	logs.DefaultLogger.LogWithTime = false
	//logs.DefaultLogger = logs.NewLogger(0)
	logs.Info("删贴机启动", time.Now())

	keepUpdatingSettings()
	logs.Info(settings)
	if settings == nil {
		return
	}

	var accAndr = accounts.NewDefaultAndroidAccount("")
	var accWin8 = accounts.NewDefaultWindows8Account("")
	accWin8.BDUSS = settings.BDUSS
	accAndr.BDUSS = settings.BDUSS

	pd := postfinder.NewPostFinder(accWin8, accAndr, settings.ForumName, func(postfinder *postfinder.PostFinder) {
		postfinder.ThreadFilter = ThreadFilter
		postfinder.NewThreadFirstAssessor = NewThreadFirstAssessor
		postfinder.NewThreadSecondAssessor = NewThreadSecondAssessor
		postfinder.AdvSearchAssessor = AdvSearchAssessor
		postfinder.PostAssessor = PostAssessor
		postfinder.CommentAssessor = CommentAssessor
	})

	if pd == nil {
		return
	}

	InitAssessor()

	pd.Run()

	<-make(chan bool)

}

var 水楼Tids map[uint64]bool
var 服务器楼Tids map[uint64]bool
var 吧规Tids map[uint64]bool
var 内容关键词 *[]*regexp.Regexp

func InitAssessor() {
	os.Mkdir("debug", 0777)
	水楼Tids = make(map[uint64]bool)
	服务器楼Tids = make(map[uint64]bool)
	吧规Tids = make(map[uint64]bool)
	内容关键词 = KeepUpdatingKeyWords(settings.PostContentRegexpFilePath)

}

func ThreadFilter(account *accounts.Account, thread *postfinder.ForumPageThread) (postfinder.Control, string) {

	if (thread.Thread.Author.Name == "MC吧饮水姬" || 包含字符串(settings.BawuList, thread.Thread.Author.Name)) &&
		strings.Contains(thread.Thread.Title, "官方水楼") {
		if thread.Thread.LastReplyer.Name == "iamunknown" {
			return postfinder.Continue, "" //测试用
		}
		水楼Tids[thread.Thread.Tid] = true
		return postfinder.Finish, "水楼"
	}

	if 包含字符串(settings.BawuList, thread.Thread.LastReplyer.Name) {
		return postfinder.Finish, "吧务发贴"
	}
	if 包含字符串(settings.BawuList, thread.Thread.Author.Name) {
		if strings.Contains(thread.Thread.Title, "服务器发布贴") {
			服务器楼Tids[thread.Thread.Tid] = true
		} else if match, _ := regexp.MatchString(`吧规.*?\([0-9]*?.*?\)`, thread.Thread.Title); match {
			吧规Tids[thread.Thread.Tid] = true
		} else if match, _ := regexp.MatchString(`基本守则`, thread.Thread.Title); match {
			吧规Tids[thread.Thread.Tid] = true
		}

	}
	return postfinder.Continue, ""
}

func AdvSearchAssessor(account *accounts.Account, result *advsearch.AdvSearchResult) postfinder.Control {

	if 水楼Tids[result.Tid] {
		if result.Author.Name == "iamunknown" {
			return postfinder.Continue //测试用
		}
		return postfinder.Finish //防止水楼回复被删
	} else if 吧规Tids[result.Tid] && !包含字符串(settings.BawuList, result.Author.Name) {
		return DeletePost("高级搜索", account, result.Tid, result.Pid, 0, 0, "非吧务回复吧规")
	}

	//DebugLog("高级搜索", result.PGetContentList())
	if len(result.Content) <= 120 {
		match, _ := regexp.MatchString(`回复.*?:`, result.Content)
		if match {
			return postfinder.Finish //疑似楼中楼而且内容完整的回复就不看了吧...
		}
	}

	if CommonAssess("高级搜索", account, result, result.Tid) == postfinder.Finish {
		return postfinder.Finish
	}

	return postfinder.Continue
}

func NewThreadFirstAssessor(account *accounts.Account, thread *postfinder.ForumPageThread) postfinder.Control {
	//DebugLog("新主题第一次", thread.Thread.TGetContentList())
	if matchedExp := 匹配正则组(*内容关键词, thread.Thread.Title); matchedExp != "" {
		return DeleteThread("主页页面", account, thread.Thread.Tid, 0, thread.Thread.Author.ID, fmt.Sprint("标题匹配关键词:", matchedExp))
	} else if matchedExp := 匹配正则组(*内容关键词, ExtractText(thread.Thread.TGetContentList())); matchedExp != "" {
		return DeleteThread("主页页面", account, thread.Thread.Tid, 0, thread.Thread.Author.ID, fmt.Sprint("内容匹配关键词:", matchedExp))
	}

	if strings.Contains(thread.Thread.Title, "乡村") &&
		!strings.Contains(thread.Thread.Title, "改造") && !strings.Contains(thread.Thread.Title, "建筑") {
		return DeleteThread("主页页面", account, thread.Thread.Tid, 0, thread.Thread.Author.ID, "乡村类垃圾主题")
	}
	if match, _ := regexp.MatchString(`(传奇.*?[A-Za-z0-9]{2}|[0-9A-Za-z]{2}.*?传奇)`, thread.Thread.Title); match {
		return DeleteThread("主页页面", account, thread.Thread.Tid, 0, thread.Thread.Author.ID, "传奇私服广告")
	}
	return postfinder.Continue
}

func NewThreadSecondAssessor(account *accounts.Account, post *postfinder.ThreadPagePost) {
	if CommonAssess("主题页面(新主题)", account, post.Post, post.Thread.Tid) == postfinder.Finish {
		return
	}

	//DebugLog("新主题第二次", post.Post.PGetContentList())

	//logs.Debug(MakePrefix(GetServerTimeFromExtra(post.Extra), post.Thread.Tid, post.Post.Pid, 0, post.Post.Author.ID),
	//	"新主题") //, post.Thread.Title, post.Post.Author, post.Post.ContentList)

}

func PostAssessor(account *accounts.Account, post *postfinder.ThreadPagePost) {
	//logs.Debug(MakePrefix(GetServerTimeFromExtra(post.Extra), post.Thread.Tid, post.Post.Pid, 0, post.Post.Author.ID),
	//	"新回复") //, post.Thread.Title, post.Post.Author, post.Post.ContentList)
	if 吧规Tids[post.Thread.Tid] && !包含字符串(settings.BawuList, post.Post.Author.Name) {
		DeletePost("主题页面", account, post.Thread.Tid, post.Post.Pid, 0, post.Post.Author.ID, "非吧务回复吧规")
		return
	}

	//DebugLog("一般回复", post.Post.PGetContentList())
	for _, content := range post.Post.ContentList {
		if link, ok := content.(postbar.Link); ok {
			if link.Text == "[语音]来自新版客户端语音功能" {
				logs.Debug("有语音")
			}
		}
	}
	if CommonAssess("主题页面", account, post.Post, post.Thread.Tid) == postfinder.Finish {
		return
	}
}

func CommentAssessor(account *accounts.Account, comment *postfinder.FloorPageComment) {
	//logs.Debug(MakePrefix(GetServerTimeFromExtra(comment.Extra), comment.Thread.Tid, comment.Post.Pid, comment.Comment.Spid, comment.Comment.Author.ID),
	//	"新楼中楼回复") //, comment.Thread.Title, comment.Post.Author, comment.Comment.Author, comment.Comment.ContentList)
	if CommonAssess("楼层页面", account, comment.Comment, comment.Thread.Tid) == postfinder.Finish {
		return
	}
	//DebugLog("楼层回复", comment.Comment.PGetContentList())

}

func CommonAssess(from string, account *accounts.Account, post postbar.IPost, tid uint64) postfinder.Control {

	_, uid := post.PGetAuthor().AGetID()
	pid := post.PGetPid()

	if 水楼Tids[tid] {
		logs.Debug(MakePrefix(nil, tid, pid, 0, uid), "水楼的贴子应该来不到这里,但是不知道为什么来了.")
		return postfinder.Finish //防止水楼回复被删
	}

	//contentList := post.PGetContentList()
	text := ExtractText(post.PGetContentList())

	if matchedExp := 匹配正则组(*内容关键词, text); matchedExp != "" {
		return DeletePost(from, account, tid, pid, 0, uid, fmt.Sprint("内容匹配关键词:", matchedExp))
	} else if math.Mod(float64(len(text)), 15.0) == 0 {
		if match, _ := regexp.MatchString("[1十拾⑩①][5五伍⑤]字", text); match {
			return DeletePost(from, account, tid, pid, 0, uid, fmt.Sprint("标准十五字", matchedExp))
		}

	}
	/*
		url, err := pbutil.GetUserSignUrl(tid, pid)
		if err == nil && url != "" {
			res, err := http.Get(url)
			if err == nil {
				exp := regexp.MustCompile(`[0-9a-z]*\....\?`)
				f, err := os.Create("debug/" + strings.TrimSuffix(exp.FindString(url), "?"))
				if err != nil {
					logs.Fatal(err)

				} else {
					io.Copy(f, res.Body)
				}
				if f != nil {
					f.Close()
				}
			} else {
				logs.Fatal(err)
			}
		}
	*/
	//fmt.Println(ExtractText(post.PGetContentList()))
	return postfinder.Continue
}

func useless() {
	fmt.Println(io.EOF,
		http.DefaultMaxHeaderBytes,
		pbutil.DataFieldNotFoundError)

}

func TidAssess(account *accounts.Account, post postbar.IPost, tid uint64) {

}

func MakePrefix(serverTime *time.Time, tid, pid, spid, uid uint64) string {
	return postfinder.MakePostLogString(serverTime, tid, pid, spid, uid)
}

func GetServerTimeFromExtra(extra postbar.IExtra) *time.Time {
	return postfinder.GetServerTimeFromExtra(extra)

}

func KeepUpdatingKeyWords(fileName string) *[]*regexp.Regexp {
	var kwExpPtr *[]*regexp.Regexp
	var lastChangeTime time.Time

	var WaitForInit chan bool = make(chan bool)

	go func() {
		var f *os.File
		var foerr, _err error
		_kwExp := make([]*regexp.Regexp, 0)
		aticker := time.NewTicker(time.Second)
		for {
			if f, foerr = os.Open(fileName); foerr != nil {
				logs.Error("更新关键词列表失败,打开文件失败.", foerr.Error())
			} else if finfo, err := f.Stat(); err != nil {
				_err = err
				logs.Error("更新关键词列表失败,获取文件信息失败.", err.Error())
			} else if finfo.ModTime() != lastChangeTime {
				lastChangeTime = finfo.ModTime()

				if keywordsBytes, err := ioutil.ReadAll(f); err != nil {
					_err = err
					logs.Error("更新关键词列表失败,读取文件失败.", err.Error())
				} else {
					logs.Info("外部关键词文件变更,更新外部关键词.")
					keywords := strings.Split(string(keywordsBytes), "\n")
					for _, v := range keywords {
						exp, err := regexp.Compile(v)
						_err = err
						if err != nil {
							logs.Error("更新关键词`"+v+"`失败.", err.Error())
						} else if exp.MatchString("") || exp.MatchString(" ") {
							logs.Error("更新关键词`"+v+"`失败.", "地图炮关键词禁止.")
						} else {
							_kwExp = append(_kwExp, exp)
							logs.Info("外部关键词:", v)
						}
					}

				}
			}
			if foerr == nil {
				f.Close()
			}
			if _err == nil && foerr == nil {
				if kwExpPtr == nil {
					kwExpPtr = &_kwExp
					WaitForInit <- true
				} else {
					*kwExpPtr = _kwExp
				}
			} else {
				logs.Error("更新关键词列表失败.重试.")
				<-aticker.C
				continue
			}
			<-aticker.C
		}

	}()
	<-WaitForInit
	close(WaitForInit)
	return kwExpPtr
}

func keepUpdatingSettings() {
	var fileName string
	if len(os.Args) == 1 {
		fileName = "删贴机设置.json"
	} else {
		fileName = os.Args[1]
	}

	var lastModTime time.Time
	ticker := time.NewTicker(time.Second)
	var isFirstTime bool = true
	var firstTimeWaitChan = make(chan bool)
	go func() {
		for {

			info, err := os.Stat(fileName)
			if err != nil {
				if isFirstTime {
					panic(err)
				}
				continue
			}

			if modTime := info.ModTime(); modTime.After(lastModTime) {
				lastModTime = modTime
				_settings, err := LoadSettings(fileName)
				if err != nil {
					logs.Fatal("更新设置文件失败,将继续使用旧设置:", err)
				} else {
					logs.Info("更新设置文件成功")
					settings = _settings
				}
			}

			if isFirstTime {
				firstTimeWaitChan <- true
				isFirstTime = false
			}

			<-ticker.C
		}

	}()

	<-firstTimeWaitChan
	close(firstTimeWaitChan)

}

func ExtractText(contentList []postbar.Content) string {
	var str string
	for _, content := range contentList {
		if text, ok := content.(postbar.Text); ok {
			str = str + text.Text + "\n"
		}
	}
	return strings.TrimSuffix(str, "\n")
}

func DebugLog(From string, contentList []postbar.Content) {
	logs.Debug(From, ":", ExtractText(contentList))
}

func 包含字符串(slice []string, sub string) bool {
	for _, str := range slice {
		if sub == str {
			return true
		}
	}
	return false
}

func 匹配正则组(exps []*regexp.Regexp, text string) string {
	for _, exp := range exps {
		if exp.MatchString(text) {
			return exp.String()
		}
	}
	return ""
}

func DeletePost(from string, account *accounts.Account, tid, pid, spid, uid uint64, reason string) postfinder.Control {

	if account.BDUSS == "" {
		logs.Fatal("BDUSS为空,忽略")
		return postfinder.Finish
	}

	var op_pid uint64
	if spid != 0 {
		op_pid = spid
	} else {
		op_pid = pid
	}

	prefix := MakePrefix(nil, tid, pid, spid, uid)
	logs.Info(prefix, from, "删贴:", reason, ".")

	for i := 0; ; i++ {
		err, pberr := apis.DeletePost(account, op_pid)
		if err == nil && (pberr == nil || pberr.ErrorCode == 0) {
			return postfinder.Finish
		} else if i < 3 {
			logs.Error(prefix, "删贴失败,将最多尝试三次:", err, pberr, ".")
		} else {
			logs.Error(prefix, "删贴失败,放弃:", err, pberr, ".")
			return postfinder.Finish
		}
	}
}

func DeleteThread(from string, account *accounts.Account, tid, pid, uid uint64, reason string) postfinder.Control {

	prefix := MakePrefix(nil, tid, pid, 0, uid)
	logs.Info(prefix, from, "删主题:", reason, ".")

	for i := 0; ; i++ {
		err, pberr := apis.DeleteThread(account, tid)
		if err == nil && (pberr == nil || pberr.ErrorCode == 0) {
			return postfinder.Finish
		} else if i < 3 {
			logs.Error(prefix, "删主题失败,将最多尝试三次:", err, pberr, ".")
		} else {
			logs.Error(prefix, "删主题失败,放弃:", err, pberr, ".")
			return postfinder.Finish
		}
	}
}
