package collects

import (
	"fmt"
	"sort"
	"time"

	"github.com/purstal/pbtools/modules/logs"
	"github.com/purstal/pbtools/modules/postbar"
	"github.com/purstal/pbtools/modules/postbar/apis/thread-win8-1.5.0.0"
)

//threads为升序, cutOffs为升序
func Validate(accWin8 *postbar.Account, threads []Thread, cutOffs []int64) [][]Thread {
	//var known []Known
	var lo = 0

	var result [][]Thread

	for i, cutOff := range cutOffs {
		var hi = len(threads) //偷懒
		x := lo + sort.Search(hi-lo, func(i int) bool {
			var j = lo + i
			var theTime int64
			for ; j >= 0; j-- {
				GetCreateTime(accWin8, &threads[j])
				if !threads[j].deleted {
					theTime = threads[j].time
					break
				}
			}
			if j == -1 {
				for j = lo + i; j < hi; j++ {
					GetCreateTime(accWin8, &threads[j])
					if !threads[j].deleted {
						theTime = threads[j].time
						break
					}
				}
			}
			return theTime >= cutOff
		})
		if x == len(cutOffs) {
			for ; i < len(cutOffs); i++ {
				result = append(result, []Thread{})
			}
			return result
		}
		logs.Debug(fmt.Sprintf("确认界限索引.界限:%s, 贴子时间:%s, 索引:%d", time.Unix(cutOff, 0).String(), time.Unix(threads[x].time, 0).String(), x))
		result = append(result, threads[lo:x])
		lo = x
	}
	result = append(result, threads[lo:])
	return result
}

func GetCreateTime(accWin8 *postbar.Account, theThread *Thread) {
	if !theThread.deleted && theThread.time == 0 {
		createTime := TryGettingThreadCreateTime(accWin8, theThread.Tid)
		if createTime == nil {
			theThread.deleted = true
		} else {
			theThread.time = createTime.Unix()
		}
	}

}

func TryGettingThreadCreateTime(accWin8 *postbar.Account, tid uint64) *time.Time {
	for {
		t, err, pberr := thread.GetThread2(accWin8, tid, false, 0, 1, 2, false, false, false)
		if err == nil {
			if pberr == nil || pberr.ErrorCode == 0 {
				if t.Thread.CreateTime.Unix() != 0 { //if len(t.PostList) != 0 {
					return &t.Thread.CreateTime
				}
				logs.Error(fmt.Sprintf("无法获取主题创造时间,重试."))
			} else if pberr.ErrorCode == 4 {
				logs.Error(fmt.Sprintf("主题已被删除,跳过. tid:%d; pberror:%d(%s).", tid, pberr.ErrorCode, pberr.ErrorMsg))
				return nil
			} else {
				logs.Error(fmt.Sprintf("无法获取主题创造时间,重试. tid:%d; pberror:%d(%s).", tid, pberr.ErrorCode, pberr.ErrorMsg))
			}
		} else {
			logs.Error(fmt.Sprintf("无法获取主题创造时间,重试. tid:%d; error:%s", tid, err.Error()))
		}
	}
}
