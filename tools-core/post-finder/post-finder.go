package post_finder

import (
	"fmt"
	//"net/http"
	"github.com/purstal/pbtools/modules/postbar"
	"github.com/purstal/pbtools/modules/postbar/accounts"
	"github.com/purstal/pbtools/modules/postbar/advsearch"
	"github.com/purstal/pbtools/modules/postbar/apis"
	"github.com/purstal/pbtools/modules/postbar/forum-win8-1.5.0.0"
	"github.com/purstal/pbtools/modules/postbar/thread-win8-1.5.0.0"
	"strconv"
	"time"
	//"github.com/purstal/pbtools/modules/logs"
	"github.com/purstal/pbtools/modules/pberrors"
	monitor "github.com/purstal/pbtools/tools-core/forum-page-monitor"
)

type Control int

const (
	Finish   Control = 0
	Continue Control = 1
)

type ThreadFilter func(account *accounts.Account, thread *ForumPageThread) (Control, string) //false则无需Continue,下同
type ThreadAssessor func(account *accounts.Account, thread *ForumPageThread) Control
type AdvSearchAssessor func(account *accounts.Account, result *advsearch.AdvSearchResult) Control
type PostAssessor func(account *accounts.Account, post *ThreadPagePost)
type CommentAssessor func(account *accounts.Account, comment *FloorPageComment)

type PostFinder struct {
	FreshPostMonitor        *monitor.FreshPostMonitor
	ServerTime              time.Time
	AccWin8, AccAndr        *accounts.Account
	ForumName               string
	Fid                     uint64
	ThreadFilter            ThreadFilter
	NewThreadFirstAssessor  ThreadAssessor
	NewThreadSecondAssessor PostAssessor
	AdvSearchAssessor       AdvSearchAssessor
	PostAssessor            PostAssessor
	CommentAssessor         CommentAssessor
	SearchTaskManager       *SearchTaskManager

	Debugger *Debugger
	Debug    struct {
		StartTime time.Time
	}
}

func init() {
	InitLoggers()
	InitDebugger()
}

func NewPostFinder(accWin8, accAndr *accounts.Account, forumName string, yield func(*PostFinder)) *PostFinder {
	var postFinder PostFinder
	postFinder.Debug.StartTime = time.Now()

	postFinder.AccWin8 = accWin8
	postFinder.AccAndr = accAndr
	postFinder.ForumName = forumName

	yield(&postFinder)
	if postFinder.ThreadFilter == nil || postFinder.NewThreadFirstAssessor == nil ||
		postFinder.NewThreadSecondAssessor == nil || postFinder.AdvSearchAssessor == nil ||
		postFinder.PostAssessor == nil || postFinder.CommentAssessor == nil {
		Logger.Fatal("删贴机初始化错误,有函数未设置:", postFinder, ".")
		panic("删贴机初始化错误,有函数未设置: " + fmt.Sprintln(postFinder) + ".")
	}

	fid, err, pberr := apis.GetFid(forumName)
	if err != nil || pberr != nil {
		Logger.Fatal("获取fid时出错: ", err, pberr)
		return nil
	}
	postFinder.Fid = fid

	postFinder.SearchTaskManager = NewSearchTaskManager(&postFinder, 0, time.Second,
		time.Second*10, time.Second*30, time.Minute, time.Minute*5, time.Minute*10,
		time.Minute*30, time.Hour, time.Hour*3)

	/*
		postFinder.SearchTaskManager = NewSearchTaskManager(&postFinder, 0, time.Second,
			time.Second*3, time.Second*3, time.Second*3)
	*/

	postFinder.Debugger = NewDebugger(forumName, &postFinder, time.Second/4)

	return &postFinder

}

type ForumPageThread struct {
	Forum  *forum.ForumPage
	Thread forum.ForumPageThread
	Extra  *forum.ForumPageExtra
}

func (pd *PostFinder) Run(monitorInterval time.Duration) {

	//monitor.MakeForumPageThreadChan(acc, "minecraft")
	var threadChan = make(chan ForumPageThread)
	pd.FreshPostMonitor = monitor.NewFreshPostMonitor(pd.AccWin8, pd.ForumName, monitorInterval)

	go func() {
		for {
			forumPage := <-pd.FreshPostMonitor.PageChan
			//fmt.Println(len(forumPage.ThreadList))
			if forumPage.Extra.ServerTime.After(pd.ServerTime) {
				pd.ServerTime = forumPage.Extra.ServerTime
			}
			fmt.Println("---", forumPage.Extra.ServerTime)
			for _, thread := range forumPage.ThreadList {
				threadChan <- ForumPageThread{
					Forum:  forumPage.Forum,
					Thread: *thread,
					Extra:  forumPage.Extra,
				}
			}
		}
	}()

	go func() {
		for {
			thread := <-threadChan
			if pd.SearchTaskManager.Debug.CurrentServerTime.Before(thread.Extra.ServerTime) {
				pd.SearchTaskManager.Debug.CurrentServerTime = thread.Extra.ServerTime
			}
			if ctrl, _ := pd.ThreadFilter(pd.AccWin8, &thread); ctrl == Continue { //true:不忽略
				//第二项是原因
				if IsNewThread(&thread.Thread) {
					go pd.FindAndAnalyseNewThread(&thread)
				} else {
					go pd.FindAndAnalyseNewPost(&thread)

				}
			} else {
				//Logger.Debug(MakePostLogString(GetServerTimeFromExtra(thread.Extra), thread.Thread.Tid, 0, 0, thread.Thread.LastReplyer.ID), "跳过: ", reason)
			}
		}
	}()

}

func (pd *PostFinder) FindAndAnalyseNewPost(thread *ForumPageThread) {
	if pd.FindAndAnalyseNewPostFromThreadPage(thread) == Continue {
		pd.SearchTaskManager.AddDemand(thread.Thread)
	}
}

func (pd *PostFinder) FindAndAnalyseNewPostFromThreadPage(forumPageThread *ForumPageThread) Control {
	var threadPage *ThreadPage = TryGettingThreadPageStruct2(pd.AccWin8, forumPageThread.Thread.Tid, false, 0, 0, 3, false, false, true)
	if threadPage == nil {
		return Continue
	}

	var HasFoundPostFromThreadPage = false

	for i, post := range threadPage.PostList {

		if post.PostTime.Unix() >= forumPageThread.Thread.LastReplyTime.Unix() {
			go func() { pd.SearchTaskManager.ReportFoundPid(post.Author.ID, post.Pid, post.PostTime) }()

			if i != 0 {
				Logger.Info(MakePostLogString(GetServerTimeFromExtra(threadPage.Extra),
					forumPageThread.Thread.Tid, post.Pid, 0, forumPageThread.Thread.LastReplyer.ID), "额外发现新回复 ")
			}
			threadPage.Thread.ForumPageThread = &forumPageThread.Thread
			pd.PostAssessor(pd.AccWin8, &ThreadPagePost{
				Thread: threadPage.Thread,
				Post:   post,
				Extra:  threadPage.Extra,
			})
			if !HasFoundPostFromThreadPage {
				if post.Author.ID == forumPageThread.Thread.LastReplyer.ID {
					HasFoundPostFromThreadPage = true
				}
			}
			//pd.AnalyseNewPostFromThreadPage(&post)
		}
	}
	if HasFoundPostFromThreadPage {
		return Finish
	} else {
		return Continue
	}
}

func (pd *PostFinder) FindAndAnalyseNewThread(forumPageThread *ForumPageThread) {
	if pd.NewThreadFirstAssessor(pd.AccWin8, forumPageThread) == Finish {
		return
	}

	var threadPage *ThreadPage = TryGettingThreadPageStruct2(pd.AccWin8, forumPageThread.Thread.Tid, false, 0, 1, 2, false, false, false)
	if threadPage == nil {

		return
	}

	floor1 := threadPage.PostList[0]
	if floor1.PostTime != forumPageThread.Thread.LastReplyTime {
		Logger.Warn(MakePostLogString(GetServerTimeFromExtra(threadPage.Extra), forumPageThread.Thread.Tid,
			floor1.Pid, 0, floor1.Author.ID), "最初判断为新主题,但不是新主题:", "时间不匹配:"+floor1.PostTime.String(),
			"!=", forumPageThread.Thread.LastReplyTime.String())
		pd.FindAndAnalyseNewPost(forumPageThread)
	} else {
		go func() { pd.SearchTaskManager.ReportFoundPid(floor1.Author.ID, floor1.Pid, floor1.PostTime) }()

		threadPage.Thread.ForumPageThread = &forumPageThread.Thread
		pd.NewThreadSecondAssessor(pd.AccWin8, &ThreadPagePost{
			Thread: threadPage.Thread,
			Post:   floor1,
			Extra:  threadPage.Extra,
		})
	}
}

type ThreadPage struct {
	Thread   *thread.ThreadPage
	PostList []thread.ThreadPagePost
	Extra    *thread.ThreadPageExtra
}

type ThreadPagePost struct {
	Thread *thread.ThreadPage
	Post   thread.ThreadPagePost
	Extra  *thread.ThreadPageExtra
}

func TryGettingThreadPageStruct(accWin8 *accounts.Account, kz uint64,
	mark bool, pid uint64, pn, rn int, withFloor, seeLz,
	r bool) (*ThreadPage, *pberrors.PbError) {
	//fmt.Println("TryGettingThreadPageStruct")
	for i, err_count := 0, 0; ; {
		tp, tpps, tpe, err, pberr := thread.GetThreadStruct(accWin8, kz, mark, pid, pn, rn, withFloor, seeLz, r)
		//fmt.Println("t", mark, err, pberr)
		threadPage := &ThreadPage{
			Thread:   tp,
			PostList: tpps,
			Extra:    tpe,
		}
		if err != nil {
			if err_count < 100 {
				err_count++
				continue
			}
			GettingStructLogger.Fatal("尝试获取主题结构无法进展,放弃.参数:", "kz=", kz, ",mark=", mark, ",pid=",
				pid, ",pn=", pn, ",rn=", rn, ",with_floor=", withFloor, ",see_lz=", seeLz, ",r=", r, ";错误:", err)
			return nil, nil

		} else if pberr == nil {
			return threadPage, nil
			/*
				} else if pberr.ErrorCode == 4 { //贴子不存在
				return nil, pberr
			*/
		} else if i < 3 {
			i++
			continue
		} else {
			return threadPage, pberr
		}
	}

}

func TryGettingThreadPageStruct2(accWin8 *accounts.Account, kz uint64,
	mark bool, pid uint64, pn, rn int, withFloor, seeLz,
	r bool) *ThreadPage {

	for j := 0; ; {
		thread, pberr := TryGettingThreadPageStruct(accWin8, kz, mark, pid, pn, rn, withFloor, seeLz, r)
		if pberr != nil {
			Logger.Error(MakePostLogString(GetServerTimeFromExtra(thread.Extra), kz, 0, 0, 0),
				"尝试获取主题时出错,放弃:", pberr, ".")
			return nil
		} else if thread == nil {
			return nil
		} else if len(thread.PostList) != 0 {
			return thread
		} else if j < 3 {
			Logger.Error(MakePostLogString(GetServerTimeFromExtra(thread.Extra), kz, 0, 0, 0),
				"返回的主题回贴列表为空,重试.")
			j++
			continue
		} else {
			Logger.Error(MakePostLogString(GetServerTimeFromExtra(thread.Extra), kz, 0, 0, 0),
				"尝试获取主题时出错,放弃:", "返回的主题回贴列表为空.")
			return nil
		}
	}

}

func MakePostLogString(serverTime *time.Time, tid, pid, spid, uid uint64) string {
	var now = time.Now()

	var localTimeStr = "L" + now.Format("15:04:05")
	var serverTimeStr string
	if serverTime != nil {
		serverTimeStr = "|S" + fmt.Sprintf("=%+d", uint64(serverTime.Sub(now).Seconds()))
	} else {
		serverTimeStr = ""
	}

	var tidStr, pidStr, uidStr string
	if tid == 0 {
		tidStr = "?"
	} else {
		tidStr = strconv.FormatUint(tid, 10)
	}
	if pid == 0 {
		//pidStr = "_"
	} else {
		pidStr = "#" + strconv.FormatUint(pid, 10)
	}
	if spid == 0 {
	} else if pid == 0 {
		pidStr = "#?." + strconv.FormatUint(spid, 10)
	} else {
		pidStr = pidStr + "." + strconv.FormatUint(spid, 10)
	}
	if uid == 0 {
		//uidStr = "_"
	} else {
		uidStr = "$" + strconv.FormatUint(uid, 10)
	}

	return "[" + localTimeStr + serverTimeStr + "|" + tidStr + pidStr + uidStr + "]"
}

func GetServerTimeFromExtra(extra postbar.IExtra) *time.Time {
	if extra != nil {
		time := extra.EGetServerTime()
		return &time
	} else {
		return nil
	}
}
