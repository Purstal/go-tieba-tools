package post_finder

import (
	//"fmt"
	"sync"
	"time"

	"github.com/purstal/pbtools/modules/postbar/forum-win8-1.5.0.0"
)

type SearchTaskManager struct {
	PostFinder *PostFinder

	TaskMap map[uint64]*SearchTask

	Intervals []time.Duration

	CurrentFoundPidsPool map[uint64][]uint64
	LastReportTime       time.Time
	LastMinFoundPidsPool map[uint64][]uint64

	TaskLock sync.Mutex

	foundPidsLock sync.Mutex

	Debug struct {
		CurrentServerTime time.Time
	}
}

type reportFoundPidStruct struct {
	Uid        uint64
	Pid        uint64
	ServerTime time.Time
}

type giveBackFoundPidsStruct struct {
	Uid  uint64
	Pids []uint64
}

func NewSearchTaskManager(deleter *PostFinder, intervals ...time.Duration) *SearchTaskManager {
	var m SearchTaskManager
	m.PostFinder = deleter
	m.Intervals = intervals
	m.CurrentFoundPidsPool = make(map[uint64][]uint64)
	m.LastMinFoundPidsPool = make(map[uint64][]uint64)

	m.TaskMap = make(map[uint64]*SearchTask)

	return &m

}

func (m *SearchTaskManager) DelayInterval(level int) time.Duration {
	if level >= len(m.Intervals) {
		return 0
	}
	return m.Intervals[level]
}

func (m *SearchTaskManager) RemoveTask(id uint64) {
	m.TaskLock.Lock()
	m.CurrentFoundPidsPool[id] = m.TaskMap[id].FoundPids
	delete(m.TaskMap, id)
	m.TaskLock.Unlock()
}

func (m *SearchTaskManager) AddDemand(thread forum.ForumPageThread) {
	m.TaskLock.Lock()
	if task, found := m.TaskMap[thread.LastReplyer.ID]; found {
		task.AddDemand(&thread)
	} else {
		m.TaskMap[thread.LastReplyer.ID] = NewSearchTask(m, &thread)
	}
	m.TaskLock.Unlock()
}

func (m *SearchTaskManager) ReportFoundPid(uid, pid uint64, serverTime time.Time) {
	m.foundPidsLock.Lock()
	if task, found := m.TaskMap[uid]; found {
		task.FoundPids = AppendUint64SliceKeepOrder(task.FoundPids, pid)
	} else {
		if serverTime.Unix()/60 > m.LastReportTime.Unix()/60 {
			m.LastMinFoundPidsPool = m.CurrentFoundPidsPool
			m.LastReportTime = serverTime
			m.CurrentFoundPidsPool = make(map[uint64][]uint64)
		}
		m.CurrentFoundPidsPool[uid] = AppendUint64SliceKeepOrder(m.CurrentFoundPidsPool[uid], pid)
	}
	m.foundPidsLock.Unlock()
}

func (m *SearchTaskManager) BorrowFoundPids(uid uint64) []uint64 {
	m.foundPidsLock.Lock()
	var pids = append(m.CurrentFoundPidsPool[uid], m.LastMinFoundPidsPool[uid]...)
	delete(m.CurrentFoundPidsPool, uid)
	delete(m.LastMinFoundPidsPool, uid)
	m.foundPidsLock.Unlock()
	return pids
}

func (m *SearchTaskManager) GiveBackFoundPids(uid uint64, pids []uint64) {
	m.foundPidsLock.Lock()
	m.CurrentFoundPidsPool[uid] = pids
	m.foundPidsLock.Unlock()
}
