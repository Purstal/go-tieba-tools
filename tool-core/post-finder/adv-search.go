package post_finder

import (
	"fmt"
	//"sync"
	"time"
	//"unsafe"

	"github.com/purstal/pbtools/modules/postbar"
	"github.com/purstal/pbtools/modules/postbar/advsearch"
	floor "github.com/purstal/pbtools/modules/postbar/apis/floor-andr-6.1.3"
	//forum "github.com/purstal/pbtools/modules/advsearch/forum-win8-1.5.0.0"
	thread "github.com/purstal/pbtools/modules/postbar/apis/thread-win8-1.5.0.0"
	//"github.com/purstal/pbtools/modules/logs"
)

/*
type AdvSearcher struct {
	SuccessPools SuccessPools
	PostFinder  *PostFinder
}

func NewAdvSearcher(PostFinder *PostFinder) *AdvSearcher {
	var searcher AdvSearcher
	searcher.PostFinder = PostFinder
	searcher.SuccessPools = MakeSuccessPools()
	return &searcher

}
*/
type FloorPageComment struct {
	Thread  *thread.ThreadPage
	Post    *thread.ThreadPagePost
	Comment floor.FloorPageComment
	Extra   *floor.FloorPageExtra
}

func (task *SearchTask) search(m *SearchTaskManager) []advsearch.AdvSearchResult {
	task.Debug.Log.Log(fmt.Sprintf("[%s|高级搜索]开始.", NowString(&task.Debug.LastLogTime)))
	defer func() {
		task.Debug.Log.Log(fmt.Sprintf("[%s|高级搜索]结束.", NowString(&task.Debug.LastLogTime)))
		task.Debug.LastSearchTime = time.Now()
	}()
	task.Debug.InSearching = true
	defer func() {
		task.Debug.InSearching = false
		task.Debug.LastSearchTime = time.Now()
	}()

	var beginTime = time.Now()
	task.Debug.Log.Log(fmt.Sprintf("[%s|高级搜索.尝试搜索]调用.参数最早需求时间:%s,最后发现的pid:%d.",
		NowString(&task.Debug.LastLogTime), task.EarliestUnfoundTime.Format("2006-01-02 15:04:05"), task.LastFoundPid))
	resultList := TryGettingAdvSearchResultListToTimeAndPid(m.PostFinder.ForumName,
		task.UserName, task.EarliestUnfoundTime, task.LastFoundPid)
	task.Debug.Log.Log(fmt.Sprintf("[%s|高级搜索.尝试搜索]耗时%s.搜索结果数:%d.",
		NowString(&task.Debug.LastLogTime), time.Now().Sub(beginTime).String(), len(resultList)))

	return resultList

}

func (task *SearchTask) DealWithResults(resultList []advsearch.AdvSearchResult, demands []Demand, foundPids []uint64, LeastTime *time.Time, m *SearchTaskManager) {
	var beginTime = time.Now()
	task.Debug.Log.Log(fmt.Sprintf("[%s|处理搜索结果]开始.", NowString(&task.Debug.LastLogTime)))
	defer func() {
		task.Debug.Log.Log(fmt.Sprintf("[%s|处理搜索结果]结束.耗时%s.", NowString(&task.Debug.LastLogTime), time.Now().Sub(beginTime).String()))
	}()

ResultIter:
	for _, result := range resultList {
		for _, pid := range foundPids {
			if pid == result.Pid {
				continue
			}
		}
		m.ReportFoundPid(task.ID, result.Pid, result.PostTime)
		if m.PostFinder.AdvSearchAssessor(m.PostFinder.AccWin8, &result) == Finish {
			continue
		}

		//先尝试楼中楼

		//var debug_pids []uint64
		floorPage, _ := TryGettingFloorPageStruct2(m.PostFinder.AccAndr, result.Tid, true, result.Pid, 0)
		if floorPage == nil {
			continue
		}

		for _, comment := range floorPage.CommentList {
			if comment.PostTime.After(*LeastTime) {
				*LeastTime = comment.PostTime
			}
			//Logger.Debug("Merry", comment.Spid, result.Pid)
			//debug_pids = append(debug_pids, comment.Spid)
			if comment.Spid == result.Pid {
				//Logger.Debug("Renko")
				for i, forumPageThread := range demands {
					if comment.PostTime == forumPageThread.LastReplyTime && floorPage.Thread.Tid == forumPageThread.Tid {
						floorPage.Thread.ForumPageThread = forumPageThread
						demands = append(demands[:i], demands[i+1:]...)
						break
					}
				}

				m.PostFinder.CommentAssessor(m.PostFinder.AccWin8, &FloorPageComment{
					Thread:  floorPage.Thread,
					Post:    floorPage.Post,
					Comment: comment,
					Extra:   floorPage.Extra,
				})
				continue ResultIter
			}
		}

		//fmt.Println("Renko")

		threadPage := TryGettingThreadPageStruct2(m.PostFinder.AccWin8, result.Tid, true, result.Pid, 0, 2, false, false, false)
		if threadPage == nil {
			continue
		}

		for _, post := range threadPage.PostList {

			if post.PostTime.After(*LeastTime) {
				*LeastTime = post.PostTime
			}
			if post.Pid == result.Pid {
				//Logger.Debug("Renko")
				for i, forumPageThread := range demands {
					if post.PostTime == forumPageThread.LastReplyTime && threadPage.Thread.Tid == forumPageThread.Tid {
						threadPage.Thread.ForumPageThread = forumPageThread
						demands = append(demands[:i], demands[i+1:]...)
						break
					}
				}
				m.PostFinder.PostAssessor(m.PostFinder.AccWin8, &ThreadPagePost{
					Thread: threadPage.Thread,
					Post:   post,
					Extra:  threadPage.Extra,
				})
				continue ResultIter
			}
		}

		Logger.Debug("高级搜索找到,楼层页面和主题页面中都未找到:", result)
	}
}

func TryGettingAdvSearchResultListToTimeAndPid(kw, un string, toTime time.Time, toPid uint64) []advsearch.AdvSearchResult {
	var allResultList []advsearch.AdvSearchResult

	var breakflag bool

	for pn := 1; ; pn++ {

		resultList := TryGettingAdvSearchResultList(kw, un, 20, pn)

		if len(resultList) == 0 {
			break
		}

		for i, result := range resultList {
			if toTime.Unix()/60 > result.PostTime.Unix()/60 || toPid >= result.Pid {
				resultList = resultList[:i]
				breakflag = true
				break
			}
		}

		if len(allResultList) != 0 {
			for i, result := range resultList {
				if result.Pid < allResultList[len(allResultList)-1].Pid {
					resultList = resultList[i:]
					break
				}
			}
		}

		allResultList = append(allResultList, resultList...)

		if breakflag {
			break
		}

	}

	return allResultList
}

func TryGettingAdvSearchResultList(kw, un string, rn,
	pn int) []advsearch.AdvSearchResult {
	for i := 0; ; {
		resultList, err := advsearch.GetAdvSearchResultList(kw, un, rn, pn)
		if len(resultList) != 0 && err == nil {
			return resultList
		} else if len(resultList) == 0 {
			if i < 3 {
				i++
			} else {
				return nil
			}
		}
	}
}

type FloorPage struct {
	Thread      *thread.ThreadPage
	Post        *thread.ThreadPagePost
	CommentList []floor.FloorPageComment
	Extra       *floor.FloorPageExtra
}

func TryGettingFloorPageStruct(accWin8 *postbar.Account, kz uint64,
	isComment bool, id uint64, pn int) (*FloorPage, *postbar.PbError) {
	for i, err_count := 0, 0; ; {
		tp, tpp, fpcs, fpe, err, pberr := floor.GetFloorStruct(accWin8, kz, isComment, id, pn)
		floorPage := &FloorPage{
			Thread:      tp,
			Post:        tpp,
			CommentList: fpcs,
			Extra:       fpe,
		}
		if err != nil {
			if err_count < 100 {
				err_count++
				continue
			}
			var pidType string = "pid"
			if isComment {
				pidType = "spid"
			}
			GettingStructLogger.Fatal("尝试获取楼层结构无法进展,放弃.参数:", "kz=", kz, ", "+pidType+"=",
				id, ",pn=", pn, ";错误:", err)

			return nil, nil
		} else if pberr == nil {
			return floorPage, nil
			/*
				} else if pberr.ErrorCode == 4 { //贴子不存在
				return floorPage, pberr
			*/
		} else if i < 3 {
			i++
			continue
		} else {
			return floorPage, pberr
		}
	}
}

func TryGettingFloorPageStruct2(accWin8 *postbar.Account, kz uint64,
	isComment bool, id uint64, pn int) (*FloorPage, bool) {
	//fmt.Println("TryGettingFloorPageStruct2")

	for i, j := 0, 0; ; {
		floorPage, pberr := TryGettingFloorPageStruct(accWin8, kz, isComment, id, pn)
		if pberr != nil {
			if i < 3 {
				//Logger.Error(MakePostLogString(GetServerTimeFromExtra(floorPage.Extra), kz, 0), "尝试获取楼层时出错,重试:", pberr, ".")
				i++
				continue
			} else {
				if pberr.ErrorCode == 4 {
					return nil, true
				} else {
					Logger.Error(MakePostLogString(GetServerTimeFromExtra(floorPage.Extra), kz, 0, 0, 0),
						"尝试获取楼层时出错,多次重试未果,放弃:", pberr, ".")
					return nil, false
				}
			}
		} else if floorPage == nil {
			return nil, false
		} else if len(floorPage.CommentList) != 0 {
			return floorPage, false
		} else if j < 3 {
			//Logger.Error(MakePostLogString(GetServerTimeFromExtra(floorPage.Extra), kz, 0), "尝试获取楼层时出错,重试:返回的楼层回贴列表为空.")
			j++
			continue
		} else {
			Logger.Error(MakePostLogString(GetServerTimeFromExtra(floorPage.Extra), kz, 0, 0, 0),
				"尝试获取楼层时出错,多次重试未果,放弃:", "返回的楼层回贴列表为空.")
			return nil, false
		}
	}
}
