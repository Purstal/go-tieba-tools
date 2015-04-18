package forum_page_monitor

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/purstal/pbtools/modules/postbar"
	"github.com/purstal/pbtools/modules/postbar/accounts"
	"github.com/purstal/pbtools/modules/postbar/forum-win8-1.5.0.0"
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

func NewFreshPostMonitor(accWin8 *accounts.Account, kw string,
	interval time.Duration) *FreshPostMonitor {

	var monitor = FreshPostMonitor{}
	monitor.forum_page_monitor = NewForumPageMonitor(accWin8, kw, interval)
	monitor.actChan = make(chan action)
	monitor.PageChan = make(chan ForumPage)

	go func() {
		var recorder = newFreshPostRecorder()

		for {
			var forumPage ForumPage
			select {
			case forumPage = <-monitor.forum_page_monitor.PageChan:
			case act := <-monitor.actChan:
				switch act.action {
				case "Stop":
					monitor.forum_page_monitor.Stop()
					close(monitor.actChan)
				}
				continue
			}

			var threadList = PageThreadsSorter(forumPage.ThreadList)
			if len(threadList) == 0 {
				continue
			}

			sort.Sort(threadList)

			var freshThreadList = make([]*forum.ForumPageThread, 0)

			var oldTime = recorder.GetOldestRecordTime()
			var thisOldestTime time.Time
			var thisNewestTime time.Time = threadList[0].LastReplyTime

			if oldTime != nil {
				thisOldestTime = threadList[0].LastReplyTime
				for _, thread := range threadList[1:] {
					if thread.LastReplyTime.Unix() < thisOldestTime.Unix() &&
						thread.LastReplyTime.Unix() >= oldTime.Unix() {
						thisOldestTime = thread.LastReplyTime
					}
				}
			} else {
				thisOldestTime = threadList[len(threadList)-1].LastReplyTime
			}

			for _, thread := range threadList {
				if oldTime != nil && thread.LastReplyTime.Before(*oldTime) {
					//fmt.Println("break", thread.LastReplyTime, *oldTime)
					break
				}
				if !thread.LastReplyTime.After(thisOldestTime) {
					if !recorder.Found(thread.LastReplyTime, thread.Tid, thread.LastReplyer.ID) {
					} else {
						continue
					}
				}
				if thread.LastReplyTime.Equal(thisNewestTime) {
					recorder.Report(thread.LastReplyTime, thread.Tid, thread.LastReplyer.ID)
				}
				freshThreadList = append(freshThreadList, thread)
				//fmt.Println(!recorder.Found(thread.LastReplyTime, thread.Tid, thread.LastReplyer.ID))
			}
			if threadCount := len(freshThreadList); threadCount != 0 {

				//fmt.Println(threadCount, monitor.forum_page_monitor.rn)
				//data, _ := json.Marshal(recorder)
				//fmt.Println(string(data))
				//fmt.Println(len(*recorder))
				//fmt.Println(freshThreadList)

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

			}

		}
	}()

	return &monitor
}

func TryGettingUserName(acc *accounts.Account, uid uint64) string {
	if uid == 0 {
		return ""
	}
	for {
		info, err := postbar.GetUserInfo(acc, uid)
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

func NewFreshPostRecorderForTest() *freshPostRecorder {
	return newFreshPostRecorder()
}

func newFreshPostRecorder() *freshPostRecorder {
	return new(freshPostRecorder)
}

const _MAX_RECORD_DAY = 3

func (recorder *freshPostRecorder) Report(postTime time.Time, tid, uid uint64) {
	var record *dayRecord
	var i int = 0
	if len(*recorder) > 0 {
		for ; i < len(*recorder); i++ {
			if postTime.After((*recorder)[i].serverTime) {
				var _recorder = *recorder
				*recorder = append((*recorder)[:i], dayRecord{postTime, make(map[uint64][]uint64)})
				if len(_recorder) >= _MAX_RECORD_DAY {
					*recorder = append(*recorder, _recorder[i:_MAX_RECORD_DAY-1]...)
				} else {
					*recorder = append(*recorder, _recorder[i:]...)
				}
				record = &(*recorder)[i]
				break
			} else if postTime.Equal((*recorder)[i].serverTime) {
				record = &(*recorder)[i]
				break
			} else {
				return
			}
		}
	} else {
		*recorder = freshPostRecorder{dayRecord{postTime, make(map[uint64][]uint64)}}
		record = &(*recorder)[0]
	}

	record.foundPostMap[tid] = append(record.foundPostMap[tid], uid)

	return

}

func (recorder *freshPostRecorder) Found(postTime time.Time, tid, uid uint64) bool {
	for i, _ := range *recorder {
		if (*recorder)[i].serverTime.Equal(postTime) {
			//fmt.Println(len((*recorder)[i].foundPostMap))
			for _, _uid := range (*recorder)[i].foundPostMap[tid] {
				if _uid == uid {
					return true
				}
			}
			return false
		}
	}
	return false
}

func (recorder *freshPostRecorder) GetOldestRecordTime() *time.Time {
	if len(*recorder) == 0 {
		return nil
	}
	return &(*recorder)[len(*recorder)-1].serverTime
}

func OldNewFreshPostMonitor(accWin8 *accounts.Account, kw string,
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
			var nowFreshPostMap = make(map[uint64]uint64)
			var nowFreshPostTime time.Time

			var threadList = PageThreadsSorter(forumPage.ThreadList)
			sort.Sort(threadList)

			var freshThreadList []*forum.ForumPageThread

			for _, thread := range threadList {
				if nowFreshPostTime.Before(thread.LastReplyTime) {
					nowFreshPostTime = thread.LastReplyTime
					freshThreadList = nil
				}
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

				fmt.Println(threadCount, monitor.forum_page_monitor.rn)

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
