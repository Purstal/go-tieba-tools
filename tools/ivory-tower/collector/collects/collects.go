package collects

import (
	//"fmt"
	"sync"
	"time"

	"github.com/purstal/go-tieba-base/logs"
	"github.com/purstal/go-tieba-base/tieba"
	"github.com/purstal/pbtools/tool-cores/utils"
)

type Thread struct {
	Title    string
	Tid      uint64
	Author   string
	Abstract []interface{}
	time     int64
	Time     string
	deleted  bool
}

func (t Thread) GetTime() int64 {
	return t.time
}

type ThreadMap map[uint64]Thread

var MAX_THREAD_NUMBER = 20
var RN = 100

func Collect(accWin8 *postbar.Account, forum string, _endTime time.Time) ThreadMap {
	var theMap = ThreadMap{}

	var endTime = &_endTime

	for i := 0; ; i++ {
		logs.Info("开始第", i+1, "轮收集")
		endTime = collect(accWin8, forum, endTime, theMap)
		//如果 endTime == nil, 表示不用再次收集
		if endTime == nil {
			logs.Info("全部收集完毕")
			return theMap
		}
		endTime = hasMissed(accWin8, forum, endTime, theMap)
		if endTime == nil {
			logs.Info("全部收集完毕")
			return theMap
		}
	}
}

func collect(accWin8 *postbar.Account, forum string, endTime *time.Time,
	theMap ThreadMap) *time.Time {

	taskManager := utils.NewUnlimitTaskManager(MAX_THREAD_NUMBER)

	var timeLock sync.Mutex

	var nextEndTime time.Time
	reportTime := func(t time.Time) {
		timeLock.Lock()
		if t.After(nextEndTime) {
			nextEndTime = t
		}
		timeLock.Unlock()
	}

	var needNextCollect bool = true

	for pn := 1; ; pn++ {
		if taskManager.IsAllTasksFinished {
			break
		}
		taskManager.DemandChan <- struct{}{}
		if !<-taskManager.DoChan {
			break
		}
		go func(pn int) {
			var threads = TryGettingForumPageThreads(accWin8, forum, RN, pn)
			if len(threads) != 0 {
				//logs.Debug(fmt.Sprintf("本页pn为%d,第一贴时间%s; 距离结束:%s", pn, threads[0].LastReplyTime.Format("2006-01-02 15:04:05"), threads[0].LastReplyTime.Sub(*endTime).String()))
			} else {
				//logs.Warn(fmt.Sprintf("经多次尝试,本页一贴没有,返回. pn:%d", pn))
				if !taskManager.IsAllTasksFinished {
					taskManager.AllTasksFinishChan <- struct{}{}
				}
				taskManager.TaskFinishesChan <- struct{}{}
				return
			}
			for i, theThread := range threads {
				reportTime(theThread.LastReplyTime)
				if theThread.LastReplyTime.Before(*endTime) {
					if theThread.IsTop || theThread.IsLivePost {
						continue
					}
					if pn == 1 && i < len(threads) {
						needNextCollect = false
					}
					if !taskManager.IsAllTasksFinished {
						taskManager.AllTasksFinishChan <- struct{}{}
					}
					taskManager.TaskFinishesChan <- struct{}{}
					return
				}
				if _, exist := theMap[theThread.Tid]; !exist {
					theMap[theThread.Tid] = Thread{
						Title:    theThread.Title,
						Tid:      theThread.Tid,
						Author:   theThread.Author.Name,
						Abstract: append(theThread.Abstract, theThread.MediaList...), /*extractAbstract(append(theThread.Abstract, theThread.MediaList...))*/
					}
				}
			}
			taskManager.TaskFinishesChan <- struct{}{}
		}(pn)
	}
	<-taskManager.WaitForFinishChan

	if needNextCollect {
		return &nextEndTime
	} else {
		return nil
	}

}

func hasMissed(accWin8 *postbar.Account, forum string, endTime *time.Time,
	theMap ThreadMap) *time.Time {
	var threads = TryGettingForumPageThreads(accWin8, forum, RN, 1)
	var nextEndTime time.Time
	for _, theThread := range threads {
		if theThread.LastReplyTime.After(nextEndTime) {
			nextEndTime = theThread.LastReplyTime
		}
		if theThread.LastReplyTime.Before(*endTime) {
			return nil
		}
		if _, exist := theMap[theThread.Tid]; !exist {
			theMap[theThread.Tid] = Thread{
				Title:    theThread.Title,
				Tid:      theThread.Tid,
				Author:   theThread.Author.Name,
				Abstract: append(theThread.Abstract, theThread.MediaList...), /*extractAbstract(append(theThread.Abstract, theThread.MediaList...))*/
			}
		}
	}
	return &nextEndTime
}
