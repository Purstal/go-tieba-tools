package caozuoliang

import (
	"sort"
	"time"
)

type Data struct {
	Bawu_un          string
	ThreadCount      int
	PostCount        int
	SameThreads      map[int]*SameThread
	SameAccounts     map[string][]*PostLog
	OldPosts         OldPosts
	Distribution     []int //以小时为单位
	Speed            []int
	RecoveredThreads []Recovered
	RecoveredPosts   []Recovered
	//1m,2m,3m,4m,5m,6m,7m,8m,9m,10m,15m,20m,25m,30m,40m,50m,1h,1.5h,2h,2.5h,3h,4h,5h,6h,12h,1d,2d,3d,5d,10d,30d

	/*
		ChThreadCount chan int
		ChPostCount   chan int
		ChSameThread  chan orzorzorzA
		ChSameAccount chan orzorzorzB
		ChOldPosts    chan orzorzorzC
	*/
}

func NewData(hour int) *Data {
	var data Data
	data.Speed = make([]int, len(SpeedClass)+1)
	data.Distribution = make([]int, hour)
	data.SameThreads = make(map[int]*SameThread)
	data.SameAccounts = make(map[string][]*PostLog)
	return &data
}

/*
type orzorzorzA struct {
	Tid int
	Ptr *PostLog
}

type orzorzorzB struct {
	Un  string
	Ptr *PostLog
}

type orzorzorzC struct {
	Ptr      *PostLog
	Interval time.Duration
}
*/
type Account struct {
	username string
}
type Post struct {
	PostType int
	Title    string
	Content  string
	Author   string
}

/*

func (data *Data) AddCount(posttype int) {
	log("AddCount:进入")
	if posttype == 主题 {
		data.ChThreadCount <- 1
	} else {
		data.ChPostCount <- 1
	}
	log("AddCount:退出")

}

func (data *Data) AddSameThread(tid int, pl *PostLog) {

	data.ChSameThread <- orzorzorzA{tid, pl}
}

func (data *Data) AddSameAccount(un string, pl *PostLog) {

	data.ChSameAccount <- orzorzorzB{un, pl}
}

func (data *Data) AddOldPosts(pl *PostLog, t time.Duration) {
	log2("收到")
	data.ChOldPosts <- orzorzorzC{pl, t}
}
*/
type SortedData struct {
	Bawu_un          string
	ThreadCount      int
	PostCount        int
	SameThreads      SameThreads
	SameAccounts     SameAccounts
	OldPosts         OldPosts
	Distribution     []int //以小时为单位
	Speed            []int
	RecoveredThreads []Recovered
	RecoveredPosts   []Recovered
	//1m,2m,3m,4m,5m,6m,7m,8m,9m,10m,15m,20m,25m,30m,40m,50m,1h,1.5h,2h,2.5h,3h,4h,5h,6h,12h,1d,2d,3d,5d,10d,30d

	/*
		ChThreadCount chan int
		ChPostCount   chan int
		ChSameThread  chan orzorzorzA
		ChSameAccount chan orzorzorzB
		ChOldPosts    chan orzorzorzC
	*/
}

type SameThreads []SameThread

type Recovered struct {
	recovertime *time.Time
	recoverby   string
	pl          *PostLog
}

type SameThread struct {
	//PostLogs                     []*PostLog
	ThreadTitle                  string
	Count                        int
	ThreadIsDeletedByTheOperator bool
}

func (t SameThreads) Len() int {
	return len(t)
}
func (t SameThreads) Less(i, j int) bool {
	if t[i].ThreadIsDeletedByTheOperator {
		if t[j].ThreadIsDeletedByTheOperator {
			return t[i].Count > t[j].Count
		}
		return true
	} else if t[j].ThreadIsDeletedByTheOperator {
		return false
	}
	return t[i].Count > t[j].Count
}
func (t SameThreads) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

type SameAccounts [][]*PostLog

func (t SameAccounts) Len() int {
	return len(t)
}
func (t SameAccounts) Less(i, j int) bool {
	return len(t[i]) > len(t[j])
}
func (t SameAccounts) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

type OldPosts []*PostLog

func (t OldPosts) Len() int {
	return len(t)
}
func (t OldPosts) Less(i, j int) bool {
	return t[i].Duration > t[j].Duration
}
func (t OldPosts) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func SortData(data *Data) *SortedData {
	var sd SortedData
	sd.Bawu_un = data.Bawu_un
	sd.PostCount = data.PostCount
	sd.ThreadCount = data.ThreadCount
	sd.Distribution = data.Distribution
	sd.Speed = data.Speed
	sd.OldPosts = data.OldPosts
	sd.RecoveredPosts = data.RecoveredPosts
	sd.RecoveredThreads = data.RecoveredThreads

	for _, v := range data.SameThreads {
		sd.SameThreads = append(sd.SameThreads, *v)
	}
	for _, v := range data.SameAccounts {
		sd.SameAccounts = append(sd.SameAccounts, []*PostLog(v))
	}
	sort.Sort(sd.SameThreads)
	sort.Sort(sd.SameAccounts)
	sort.Sort(sd.OldPosts)

	return &sd
}
