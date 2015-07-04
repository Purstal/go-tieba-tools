package post_finder

import (
	"fmt"
	"sync"
	"time"

	"github.com/purstal/pbtools/modules/postbar/apis/forum-win8-1.5.0.0"
)

type Demand *forum.ForumPageThread

type SearchTask struct {
	//Manager *SearchTaskManager

	//LastSearchTime time.Time
	Level int
	Timer *time.Timer

	ID       uint64
	UserName string

	Demands             []Demand
	FoundPids           []uint64
	LastFoundPid        uint64
	EarliestUnfoundTime time.Time

	demandLock sync.Mutex

	Debug struct {
		InSearching    bool
		LastSearchTime time.Time
		Log            debugTaskLogs //debugTaskLogs
		LastLogTime    time.Time
	}
}

type getDemandsRequestStruct struct {
	Pid        uint64
	TimeMinute time.Time
}

type returnDemandsStruct struct {
	Demands []Demand
	Pids    []uint64
}

func NewSearchTask(manager *SearchTaskManager, firstDemand Demand) *SearchTask {
	var task SearchTask

	//task.Manager = manager

	task.ID = firstDemand.LastReplyer.ID
	task.UserName = firstDemand.LastReplyer.Name

	task.Demands = []Demand{firstDemand}
	task.EarliestUnfoundTime = firstDemand.LastReplyTime

	task.FoundPids = manager.BorrowFoundPids(task.ID)

	task.Timer = time.NewTimer(manager.DelayInterval(0)) //最小单位

	go func() {
		for {
			<-task.Timer.C
			task.Debug.Log.Log(fmt.Sprintf("[%s|处理任务]开始.阶段:%d;需求数:%d.", NowString(&task.Debug.LastLogTime), task.Level, len(task.Demands)))

			resultList := task.search(manager)
			if len(resultList) != 0 {
				var leastTime = task.EarliestUnfoundTime
				demands, foundPids := task.GetDemands(resultList[0].Pid, resultList[0].PostTime)
				task.DealWithResults(resultList, demands, foundPids, &leastTime, manager)
				task.removeDemand(leastTime)
				if len(task.Demands) == 0 {
					manager.RemoveTask(task.ID)
					return
				}
			}
			task.Level++
			interval := manager.DelayInterval(task.Level)
			if interval != 0 {
				task.Timer.Reset(interval)
			} else {
				manager.RemoveTask(task.ID)
				logger.Error("放弃高级搜索任务:", task.ID)
				return
			}
			task.Debug.Log.Log(fmt.Sprintf("[%s|处理任务]结束.剩余需求数:%d;下次进行间隔:%s.",
				NowString(&task.Debug.LastLogTime), len(task.Demands), interval.String()))

		}
		defer func() {
			manager.GiveBackFoundPids(task.ID, task.FoundPids)
		}()
	}()

	return &task
}

func (task *SearchTask) AddDemand(demand Demand) {
	task.demandLock.Lock()
	if task.EarliestUnfoundTime.After(demand.LastReplyTime) {
		task.EarliestUnfoundTime = demand.LastReplyTime
	}
	task.Level = 0
	task.Timer.Reset(1)
	task.Demands = append(task.Demands, demand)
	task.demandLock.Unlock()
}

func (task *SearchTask) GetDemands(pid uint64,
	timeMinute time.Time) (
	[]Demand, []uint64) {
	task.demandLock.Lock()
	task.Debug.Log = append(task.Debug.Log, fmt.Sprintf("[%s|获取需求]参数pid:%d,参数限定时间:%s;现存需求:%d,现存发现pid:%d.",
		NowString(&task.Debug.LastLogTime), pid, timeMinute.Format("2006-01-02 15:04"), len(task.Demands), len(task.FoundPids)))
	var ret returnDemandsStruct
	task.LastFoundPid = pid
	var pids []uint64
	var x bool
	for i, foundPid := range task.FoundPids {
		if foundPid > pid {
			pids = task.FoundPids[i:]
			task.FoundPids = task.FoundPids[:i]
			x = true
			break
		}
	}
	if !x {
		ret.Pids = task.FoundPids
		task.FoundPids = nil
	}

	var demands []Demand
	for _, demand := range task.Demands {
		if demand.LastReplyTime.Unix()/60 <= timeMinute.Unix()/60 {
			demands = append(ret.Demands, demand)
		}
	}
	task.Debug.Log = append(task.Debug.Log, fmt.Sprintf("[%s|获取需求]返回需求数量:%d,返回发现pid数量:%d;剩余发现pid数量:%d.",
		NowString(&task.Debug.LastLogTime), len(demands), len(pids), len(task.FoundPids)))
	task.demandLock.Unlock()
	return demands, pids
}

func (task *SearchTask) removeDemand(leastTime time.Time) {
	task.demandLock.Lock()
	var newDemands []Demand
	for _, demand := range task.Demands {
		if demand.LastReplyTime.Unix() > leastTime.Unix() {
			newDemands = append(newDemands, demand)
		}
	}
	task.Demands = newDemands
	task.demandLock.Unlock()
}

func NowString(lastLogTime *time.Time) string {
	var now = time.Now()
	if lastLogTime.Day() != now.Day() || lastLogTime.Month() != now.Month() || lastLogTime.Year() != now.Year() {
		*lastLogTime = now
		return time.Now().Format("2006-01-02 15:04:05")
	} else {
		*lastLogTime = now
		return time.Now().Format("15:04:05")
	}

}
