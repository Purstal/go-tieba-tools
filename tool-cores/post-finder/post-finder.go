package post_finder

import (
	"fmt"
	"time"

	"github.com/purstal/go-tieba-base/tieba"
	"github.com/purstal/go-tieba-base/tieba/adv-search"
	"github.com/purstal/go-tieba-base/tieba/apis"
	"github.com/purstal/go-tieba-base/tieba/apis/forum-win8-1.5.0.0"
	monitor "github.com/purstal/pbtools/tool-cores/forum-page-monitor"
)

type Control int

const (
	Finish   Control = 0
	Continue Control = 1
)

type ThreadAssessor func(account *postbar.Account, thread *ForumPageThread) Control
type AdvSearchAssessor func(account *postbar.Account, result *advsearch.AdvSearchResult) Control
type PostAssessor func(account *postbar.Account, post *ThreadPagePost)
type CommentAssessor func(account *postbar.Account, comment *FloorPageComment)

type PostFinder struct {
	FreshPostMonitor        *monitor.FreshPostMonitor
	ServerTime              time.Time
	AccWin8, AccAndr        *postbar.Account
	ForumName               string
	Fid                     uint64
	ThreadFilter            ThreadAssessor
	NewThreadFirstAssessor  ThreadAssessor
	NewThreadSecondAssessor PostAssessor
	AdvSearchAssessor       AdvSearchAssessor
	PostAssessor            PostAssessor
	CommentAssessor         CommentAssessor
	SearchTaskManager       *SearchTaskManager

	//Abandon

	Debugger *Debugger
	Debug    struct {
		StartTime time.Time
	}
}

func NewPostFinder(accWin8, accAndr *postbar.Account, forumName string, yield func(*PostFinder), debug bool, logDir string) (*PostFinder, error) {
	var postFinder PostFinder
	postFinder.Debug.StartTime = time.Now()

	postFinder.AccWin8 = accWin8
	postFinder.AccAndr = accAndr
	postFinder.ForumName = forumName

	initLoggers(&postFinder, logDir)

	yield(&postFinder)
	if postFinder.ThreadFilter == nil || postFinder.NewThreadFirstAssessor == nil ||
		postFinder.NewThreadSecondAssessor == nil || postFinder.AdvSearchAssessor == nil ||
		postFinder.PostAssessor == nil || postFinder.CommentAssessor == nil {
		logger.Fatal("删贴机初始化错误,有函数未设置:", postFinder, ".")
		panic("删贴机初始化错误,有函数未设置: " + fmt.Sprintln(postFinder) + ".")
	}

	fid, err, pberr := apis.GetFid(forumName)
	if err != nil || pberr != nil {
		logger.Fatal("获取fid时出错: ", err, pberr)
		return nil, err
	}
	postFinder.Fid = fid

	postFinder.SearchTaskManager = NewSearchTaskManager(&postFinder, 0, time.Second,
		time.Second*10, time.Second*30, time.Minute, time.Minute*5, time.Minute*10,
		time.Minute*30, time.Hour, time.Hour*3)

	if debug {
		InitDebugger()
		postFinder.Debugger = NewDebugger(forumName, &postFinder, time.Second/4)
	}

	return &postFinder, nil

}

type ForumPageThread struct {
	Forum  *forum.ForumPage
	Thread forum.ForumPageThread
	Extra  *forum.ForumPageExtra
}

func (finder *PostFinder) Run(monitorInterval time.Duration) {

	var threadChan = make(chan ForumPageThread)
	finder.FreshPostMonitor = monitor.NewFreshPostMonitor(finder.AccWin8, finder.ForumName, monitorInterval)

	go func() {
		for {
			forumPage := <-finder.FreshPostMonitor.PageChan
			if forumPage.Extra.ServerTime.After(finder.ServerTime) {
				finder.ServerTime = forumPage.Extra.ServerTime
			}
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
			if finder.SearchTaskManager.Debug.CurrentServerTime.Before(thread.Extra.ServerTime) {
				finder.SearchTaskManager.Debug.CurrentServerTime = thread.Extra.ServerTime
			}
			if ctrl := finder.ThreadFilter(finder.AccWin8, &thread); ctrl == Continue {
				if IsNewThread(&thread.Thread) {
					go finder.FindAndAnalyseNewThread(&thread)
				} else {
					go finder.FindAndAnalyseNewPost(&thread)
				}
			}
		}
	}()
}

func (finder *PostFinder) ChangeMonitorInterval(monitorInterval time.Duration) {
	finder.FreshPostMonitor.ChangeInterval(monitorInterval)
}
