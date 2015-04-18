package post_finder

import (
	//"bytes"
	"fmt"
	//"github.com/purstal/pbtools/from_go/encoding/json"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	//"github.com/pquerna/ffjson/ffjson"
)

type DebugErrorStruct struct {
	DebugError struct {
		ErrorMessage string
	}
}

var TaskNotFoundErrorJson []byte
var WrongParametersErrorJson []byte

func InitDebugger() {

	var (
		TaskNotFoundError,
		WrongParametersError DebugErrorStruct
	)

	TaskNotFoundError.DebugError.ErrorMessage = "任务未找到"
	TaskNotFoundErrorJson, _ = json.Marshal(TaskNotFoundError)
	WrongParametersError.DebugError.ErrorMessage = "参数错误"
	WrongParametersErrorJson, _ = json.Marshal(WrongParametersError)
}

type simpleCache struct {
	LastJson        []byte
	LastCollectTime time.Time
}

type Debugger struct {
	finder      *PostFinder
	minInterval time.Duration

	tasksCache         simpleCache
	foundPidsPoolCache simpleCache
}

func NewDebugger(name string, Finder *PostFinder, minInterval time.Duration) *Debugger {
	var debugger Debugger
	debugger.finder = Finder
	debugger.minInterval = minInterval
	debuggerURI := "/debug/" + name + "/"
	http.HandleFunc(debuggerURI+"json/tasks", debugger.GetTasks)
	http.HandleFunc(debuggerURI+"json/task/logs", debugger.GetTaskLogs)
	http.HandleFunc(debuggerURI+"json/task/demands", debugger.GetTaskDemands)
	http.HandleFunc(debuggerURI+"json/found-pids-pool", debugger.GetFoundPidsPool)
	Logger.Debug("debugger:", name, debuggerURI+"json/tasks")

	return &debugger
}

func (debugger *Debugger) ChangeMinInterval(minInterval time.Duration) {
	debugger.minInterval = minInterval
}

func (debugger *Debugger) GetTasks(rw http.ResponseWriter, req *http.Request) {

	if now := time.Now(); now.Sub(debugger.tasksCache.LastCollectTime).Seconds() > debugger.minInterval.Seconds() {
		common := debugger.collectCommon()
		tasks := debugger.collectTask(common)
		var json, _ = json.MarshalIndent(tasks, "", "  ")
		debugger.tasksCache = simpleCache{json, now}
		rw.Write(json)
	} else {
		rw.Write(debugger.tasksCache.LastJson)
	}
}

func (debugger *Debugger) GetFoundPidsPool(rw http.ResponseWriter, req *http.Request) {

	if now := time.Now(); now.Sub(debugger.foundPidsPoolCache.LastCollectTime).Seconds() > debugger.minInterval.Seconds() {
		common := debugger.collectCommon()
		foundPidsPool := debugger.collectFoundPidsPool(common)
		var json, _ = json.MarshalIndent(foundPidsPool, "", "  ")
		debugger.foundPidsPoolCache = simpleCache{json, now}
		rw.Write(json)
	} else {
		rw.Write(debugger.foundPidsPoolCache.LastJson)
	}
}

func (debugger *Debugger) Collect() {

}

type debugCommon struct {
	StartTime  string `json:"启动时间"`
	Duration   string `json:"运行间隔"`
	localTime  time.Time
	LocalTime  string `json:"本地时间"`
	ServerTime string `json:"贴吧服务器时间"`

	ForumName string `json:"贴吧名"`
}

func (debugger *Debugger) collectCommon() *debugCommon {
	finder := debugger.finder
	var common debugCommon
	common.StartTime = finder.Debug.StartTime.Format("2006-01-02 15:04:05")
	common.localTime = time.Now()
	common.LocalTime = common.localTime.Format("2006-01-02 15:04:05.0000")
	common.Duration = common.localTime.Sub(finder.Debug.StartTime).String()
	common.ServerTime = finder.ServerTime.Format("2006-01-02 15:04:05")
	common.ForumName = finder.ForumName
	return &common
}

type debugTasks struct {
	Common *debugCommon // `json:"基本信息"`

	TotalTaskCount         int
	InSearchingTaskCount   int
	TotalDemandCount       int
	InSearchingDemandCount int

	inLevel [][2][2]int
	InLevel string `json:"阶段计数[[任务 搜索中任务][需求 搜索中需求]]"`

	TasksInSearching []*debugTask
	TasksInWaiting   []*debugTask
}

type debugTask struct {
	UserID   uint64
	UserName string

	Level int

	LastSearchTime     string
	WaitInterval       string
	RemainWaitInterval string

	FoundPidCount int

	LogCount int
	//Log debugTaskLogs

	DemandCount int
	//Demands     []debugDemand
}

type debugDemand struct {
	Tid      uint64
	PostTime string
}

func (debugger *Debugger) collectTask(common *debugCommon) *debugTasks {
	m := debugger.finder.SearchTaskManager

	var dtasks debugTasks
	dtasks.Common = common
	dtasks.inLevel = make([][2][2]int, len(m.Intervals))

	dtasks.TasksInSearching, dtasks.TasksInWaiting = make([]*debugTask, 0), make([]*debugTask, 0)
	taskMap := m.TaskMap

	for _, task := range taskMap {
		var dtask debugTask
		dtasks.TotalTaskCount++
		demandCount := len(task.Demands)
		dtasks.TotalDemandCount += demandCount
		inSearching := task.Debug.InSearching
		if inSearching {
			dtasks.InSearchingTaskCount++
			dtasks.InSearchingDemandCount += demandCount
		}
		level := task.Level
		if level < len(m.Intervals) {
			dtasks.inLevel[level][0][0]++
			dtasks.inLevel[level][1][0] += demandCount
			if inSearching {
				dtasks.inLevel[level][0][1]++
				dtasks.inLevel[level][1][1] += demandCount
			}
		}

		dtask.UserID = task.ID
		dtask.UserName = task.UserName
		dtask.Level = level

		dtask.LastSearchTime = task.Debug.LastSearchTime.Format("2006-01-02 15:04:05.0000")
		if !inSearching {
			waitInterval := m.DelayInterval(level)
			dtask.WaitInterval = waitInterval.String()
			dtask.RemainWaitInterval = task.Debug.LastSearchTime.Add(waitInterval).Sub(common.localTime).String()
		} else {
			dtask.WaitInterval = "N/A"
			dtask.RemainWaitInterval = "N/A"
		}

		dtask.FoundPidCount = len(task.FoundPids)

		//dtask.Log = task.Debug.Log

		dtask.LogCount = len(task.Debug.Log)
		dtask.DemandCount = demandCount
		/*
			dtask.Demands = make([]debugDemand, 0)
			for _, demand := range task.Demands {
				dtask.Demands = append(dtask.Demands, debugDemand{demand.Tid, demand.LastReplyTime.Format("2006-01-02 15:04:05")})
			}
		*/
		if inSearching {
			dtasks.TasksInSearching = append(dtasks.TasksInSearching, &dtask)
		} else {
			dtasks.TasksInWaiting = append(dtasks.TasksInWaiting, &dtask)
		}
	}

	dtasks.InLevel = fmt.Sprint(dtasks.inLevel)

	return &dtasks

}

type debugFoundPidsPool struct {
	Common *debugCommon

	TotalFoundPidCount, CurrentFoundPidCount, LastMinFoundPidCount    int
	TotalFoundUserCount, CurrentFoundUserCount, LastMinFoundUserCount int

	CurrentFoundPidsPool, LastMinFoundPidsPool map[string]debugPidsPoolReport
}

type debugPidsPoolReport struct {
	PidCount int
	Pids     string
}

func (debugger *Debugger) collectFoundPidsPool(common *debugCommon) *debugFoundPidsPool {
	currentFoundPidsPool, lastMinFoundPidsPool :=
		debugger.finder.SearchTaskManager.CurrentFoundPidsPool,
		debugger.finder.SearchTaskManager.LastMinFoundPidsPool

	var foundPidsPool debugFoundPidsPool
	foundPidsPool.Common = common

	foundPidsPool.CurrentFoundPidsPool = make(map[string]debugPidsPoolReport)
	var currentFoundPidsCount int
	for _key, value := range currentFoundPidsPool {
		var key = strconv.FormatUint(_key, 10)
		pool := debugPidsPoolReport{}
		pool.Pids = fmt.Sprint(value)
		pool.PidCount = len(value)
		currentFoundPidsCount += pool.PidCount
		foundPidsPool.CurrentFoundPidsPool[key] = pool
	}
	foundPidsPool.LastMinFoundPidsPool = make(map[string]debugPidsPoolReport)
	var lastMinFoundPidsCount int
	for _key, value := range lastMinFoundPidsPool {
		var key = strconv.FormatUint(_key, 10)
		pool := debugPidsPoolReport{}
		pool.Pids = fmt.Sprint(value)
		pool.PidCount = len(value)
		lastMinFoundPidsCount += pool.PidCount
		foundPidsPool.LastMinFoundPidsPool[key] = pool
	}
	foundPidsPool.CurrentFoundUserCount = len(foundPidsPool.CurrentFoundPidsPool)
	foundPidsPool.LastMinFoundUserCount = len(foundPidsPool.LastMinFoundPidsPool)
	foundPidsPool.TotalFoundUserCount = foundPidsPool.CurrentFoundUserCount + foundPidsPool.LastMinFoundUserCount
	foundPidsPool.CurrentFoundPidCount = currentFoundPidsCount
	foundPidsPool.LastMinFoundPidCount = lastMinFoundPidsCount
	foundPidsPool.TotalFoundPidCount = currentFoundPidsCount + lastMinFoundPidsCount

	return &foundPidsPool
}

func MakeSimpleHandler(data *[]byte) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		rw.Write(*data)
	}
}

type TaskLogCache struct {
	LogJson     []byte
	TaskID      uint64
	LogLen      int
	LastHitTime int64
}

var taskLogCaches [20]*TaskLogCache

func (debugger *Debugger) GetTaskLogs(rw http.ResponseWriter, req *http.Request) {
	var taskID, _ = strconv.ParseUint(req.FormValue("uid"), 10, 64)
	if taskID == 0 {
		rw.Write(WrongParametersErrorJson)
		return
	}
	if task, found := debugger.finder.SearchTaskManager.TaskMap[taskID]; found {
		var earliestCacheIndex int
		for i, cache := range taskLogCaches {
			if cache == nil {
				taskLogCaches[i] = new(TaskLogCache)
				continue
			}
			if taskID == cache.TaskID && len(task.Debug.Log) == cache.LogLen {
				cache.LastHitTime = time.Now().Unix()
				rw.Write(cache.LogJson)
				return
			}
			if taskLogCaches[earliestCacheIndex].LastHitTime > cache.LastHitTime {
				earliestCacheIndex = i
			}
		}
		var logJson, err = json.Marshal(task.Debug.Log)
		if err != nil {
			var MarshalError DebugErrorStruct
			MarshalError.DebugError.ErrorMessage = err.Error()
			MarshalErrorJson, _ := json.Marshal(err)
			rw.Write(MarshalErrorJson)
		} else {
			var cache = &TaskLogCache{
				LogJson:     logJson,
				TaskID:      taskID,
				LogLen:      len(task.Debug.Log),
				LastHitTime: time.Now().Unix(),
			}
			taskLogCaches[earliestCacheIndex] = cache
			rw.Write(logJson)
		}
	} else {
		rw.Write(TaskNotFoundErrorJson)
	}
}

func (debugger *Debugger) GetTaskDemands(rw http.ResponseWriter, req *http.Request) {
	var taskID, _ = strconv.ParseUint(req.FormValue("uid"), 10, 64)
	if taskID == 0 {
		rw.Write(WrongParametersErrorJson)
		return
	}
	if task, found := debugger.finder.SearchTaskManager.TaskMap[taskID]; found {
		var demands = make([]debugDemand, 0)
		for _, demand := range task.Demands {
			demands = append(demands, debugDemand{demand.Tid, demand.LastReplyTime.Format("2006-01-02 15:04:05")})
		}
		var demandsJson, err = json.Marshal(demands)
		if err != nil {
			var MarshalError DebugErrorStruct
			MarshalError.DebugError.ErrorMessage = err.Error()
			MarshalErrorJson, _ := json.Marshal(err)
			rw.Write(MarshalErrorJson)
		} else {
			rw.Write(demandsJson)
		}

	} else {
		rw.Write(TaskNotFoundErrorJson)
	}
}
