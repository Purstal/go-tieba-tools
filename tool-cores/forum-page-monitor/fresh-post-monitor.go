package forum_page_monitor

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/purstal/go-tieba-base/tieba"
	"github.com/purstal/go-tieba-base/tieba/apis"
	"github.com/purstal/go-tieba-base/tieba/apis/forum-win8-1.5.0.0"
)

type FreshPostMonitor struct {
	forum_page_monitor *ForumPageMonitor
	actChan            chan action
	PageChan           chan ForumPage
}

func useless() {
	fmt.Println(json.Marshal(nil))
}

func (filter *FreshPostMonitor) ChangeInterval(newInterval time.Duration) {
	filter.forum_page_monitor.ChangeInterval(newInterval)
}

func (filter *FreshPostMonitor) Stop() {
	filter.actChan <- action{"Stop", nil}
}

func TryGettingUserName(acc *postbar.Account, uid uint64) string {
	if uid == 0 {
		return ""
	}
	for {
		info, err := apis.GetUserInfo(acc, uid)
		if err == nil {
			switch info.ErrorCode.(type) {
			case (float64):
				if info.ErrorCode.(float64) == 0 {
					return info.User.Name
				}
			case (string):
				if info.ErrorCode.(string) == "0" {
					return info.User.Name
				}
			}
		}
	}
}

type freshPostRecorder []dayRecord

type dayRecord struct {
	serverTime   time.Time
	foundPostMap map[uint64][]uint64 //tid->uids
}

func NewFreshPostMonitor(accWin8 *postbar.Account, kw string,
	interval time.Duration) *FreshPostMonitor {

	var monitor = FreshPostMonitor{}
	monitor.forum_page_monitor = NewForumPageMonitor(accWin8, kw, interval)
	monitor.actChan = make(chan action)
	monitor.PageChan = make(chan ForumPage)

	go func() {
		var serverTimeNowInt int64
		var logIDNow uint64
		var lastFreshPostMap = make(map[uint64]uint64)
		var lastFreshPostTime time.Time //上次新回复的时间

		for {
			var forumPage ForumPage
			select {
			case forumPage = <-monitor.forum_page_monitor.PageChan:
			case act := <-monitor.actChan:
				switch act.action {
				case "stop":
					monitor.forum_page_monitor.Stop()
					close(monitor.actChan)
				}
				continue
			}

			if forumPage.Extra.ServerTime.Unix() < serverTimeNowInt {
				fmt.Println("响应的服务器时间早于上次,忽略")
				continue
			} else if forumPage.Extra.ServerTime.Unix() == serverTimeNowInt && forumPage.Extra.LogID < logIDNow {
				fmt.Println("响应的服务器时间等于上次,且log_id小于上次,忽略")
				continue
			} else {
				logIDNow = forumPage.Extra.LogID
				serverTimeNowInt = forumPage.Extra.ServerTime.Unix()
			}

			var threadList = PageThreadsSorter(forumPage.ThreadList)
			sort.Sort(threadList)

			var nowFreshPostTime time.Time
			for _, thread := range threadList {
				if thread.LastReplyTime.Before(lastFreshPostTime) {
					if thread.IsTop {
						continue
					}
					break
				}
				if nowFreshPostTime.Before(thread.LastReplyTime) {
					nowFreshPostTime = thread.LastReplyTime
				}
			}

			var nowFreshPostMap map[uint64]uint64
			if nowFreshPostTime.Equal(lastFreshPostTime) {
				nowFreshPostMap = lastFreshPostMap
			} else {
				nowFreshPostMap = make(map[uint64]uint64)
			}

			var freshThreadList []*forum.ForumPageThread
			for _, thread := range threadList {
				if thread.LastReplyTime.Before(lastFreshPostTime) {
					if thread.IsTop {
						continue
					}
					break
				} else if thread.LastReplyTime.Equal(lastFreshPostTime) {
					if lastFreshPostMap[thread.Tid] == thread.LastReplyer.ID {
						continue
					}
				}
				if thread.LastReplyTime.Equal(nowFreshPostTime) {
					nowFreshPostMap[thread.Tid] = thread.LastReplyer.ID
				}

				if thread.LastReplyer.Name == "" {
					thread.LastReplyer.Name = TryGettingUserName(accWin8, thread.LastReplyer.ID)
				}
				freshThreadList = append(freshThreadList, thread)

			}
			if threadCount := len(freshThreadList); threadCount != 0 {

				//fmt.Println(threadCount, monitor.forum_page_monitor.rn)

				if newRn := 8 + threadCount*2; newRn > 100 {
					monitor.forum_page_monitor.rn = 100
				} else {
					monitor.forum_page_monitor.rn = 8 + threadCount*2
				}

				monitor.PageChan <- ForumPage{
					Forum:      forumPage.Forum,
					ThreadList: freshThreadList,
					Extra:      forumPage.Extra,
				}

				lastFreshPostTime = nowFreshPostTime
				lastFreshPostMap = nowFreshPostMap

			}

		}
	}()

	return &monitor
}

func NewFreshPostMonitorAdv(accWin8 *postbar.Account, kw string,
	interval time.Duration, fn func(ForumPage)) *FreshPostMonitor {

	var monitor = FreshPostMonitor{}
	monitor.forum_page_monitor = NewForumPageMonitor(accWin8, kw, interval)
	monitor.actChan = make(chan action)
	monitor.PageChan = make(chan ForumPage)

	go func() {
		var serverTimeNowInt int64
		var logIDNow uint64
		var lastFreshPostMap = make(map[uint64]uint64)
		var lastFreshPostTime time.Time //上次新回复的时间

		for {
			var forumPage ForumPage
			select {
			case forumPage = <-monitor.forum_page_monitor.PageChan:
				fn(forumPage)
			case act := <-monitor.actChan:
				switch act.action {
				case "stop":
					monitor.forum_page_monitor.Stop()
					close(monitor.actChan)
				}
				continue
			}

			if forumPage.Extra.ServerTime.Unix() < serverTimeNowInt {
				fmt.Println("响应的服务器时间早于上次,忽略")
				continue
			} else if forumPage.Extra.ServerTime.Unix() == serverTimeNowInt && forumPage.Extra.LogID < logIDNow {
				fmt.Println("响应的服务器时间等于上次,且log_id小于上次,忽略")
				continue
			} else {
				logIDNow = forumPage.Extra.LogID
				serverTimeNowInt = forumPage.Extra.ServerTime.Unix()
			}

			var threadList = PageThreadsSorter(forumPage.ThreadList)
			sort.Sort(threadList)

			var nowFreshPostTime time.Time
			for _, thread := range threadList {
				if thread.LastReplyTime.Before(lastFreshPostTime) {
					if thread.IsTop {
						continue
					}
					break
				}
				if nowFreshPostTime.Before(thread.LastReplyTime) {
					nowFreshPostTime = thread.LastReplyTime
				}
			}

			var nowFreshPostMap map[uint64]uint64
			if nowFreshPostTime.Equal(lastFreshPostTime) {
				nowFreshPostMap = lastFreshPostMap
			} else {
				nowFreshPostMap = make(map[uint64]uint64)
			}

			var freshThreadList []*forum.ForumPageThread
			for _, thread := range threadList {
				if thread.LastReplyTime.Before(lastFreshPostTime) {
					if thread.IsTop {
						continue
					}
					break
				} else if thread.LastReplyTime.Equal(lastFreshPostTime) {
					if lastFreshPostMap[thread.Tid] == thread.LastReplyer.ID {
						continue
					}
				}
				if thread.LastReplyTime.Equal(nowFreshPostTime) {
					nowFreshPostMap[thread.Tid] = thread.LastReplyer.ID
				}

				if thread.LastReplyer.Name == "" {
					thread.LastReplyer.Name = TryGettingUserName(accWin8, thread.LastReplyer.ID)
				}
				freshThreadList = append(freshThreadList, thread)

			}
			if threadCount := len(freshThreadList); threadCount != 0 {

				//fmt.Println(threadCount, monitor.forum_page_monitor.rn)

				if newRn := 8 + threadCount*2; newRn > 100 {
					monitor.forum_page_monitor.rn = 100
				} else {
					monitor.forum_page_monitor.rn = 8 + threadCount*2
				}

				monitor.PageChan <- ForumPage{
					Forum:      forumPage.Forum,
					ThreadList: freshThreadList,
					Extra:      forumPage.Extra,
				}

				lastFreshPostTime = nowFreshPostTime
				lastFreshPostMap = nowFreshPostMap

			}

		}
	}()

	return &monitor
}
