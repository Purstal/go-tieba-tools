package main

import (
	"fmt"
	"time"

	"github.com/purstal/pbtools/modules/pberrors"
	"github.com/purstal/pbtools/modules/postbar"
	"github.com/purstal/pbtools/modules/postbar/thread-win8-1.5.0.0"
)

var acc = postbar.NewDefaultWindows8Account("")

func main() {

	fromTime := time.Date(2012, 8, 20, 23, 57, 0, 0, time.Local)
	//2012-08-20 23:57
	toTime := time.Date(2013, 4, 2, 9, 23, 0, 0, time.Local)
	//2012-10-11 02:19
	CollectPost(1766018024, fromTime, toTime)

}

type ThreadX struct {
	Info struct {
		Author string
		Title  string
	}
	PostList []thread.ThreadPagePost
}

func CollectPost(tid uint64, from, to time.Time) ThreadX {
	var tx ThreadX
	firstPage, pberr := tryGettingThreadStruct(tid, 1)

	if pberr != nil && pberr.ErrorCode != 0 {
		fmt.Println("无法获取主题", tid, "第一页:", pberr, ",放弃.")
		return tx
	}

	//var postsInRange []thread.ThreadPagePost
	var /*postsInRangeFirstPage,*/ postsInRangeLastPage []thread.ThreadPagePost
	var noMore bool

	if tx.PostList, noMore = ExtractPostInRange(firstPage.PostList, from, to, false); noMore {
		return tx
	}

	lastPage, _ := tryGettingThreadStruct(tid, firstPage.Extra.TotalPage)
	if postsInRangeLastPage, noMore = ExtractPostInRange(lastPage.PostList, from, to, true); noMore {
		tx.PostList = postsInRangeLastPage
		return tx
	}

	if len(tx.PostList) == 0 {
		firstPage = findPageInRange(tid, firstPage.Extra.TotalPage, from, false)
		if tx.PostList, noMore = ExtractPostInRange(firstPage.PostList, from, to, false); noMore {
			return tx
		}
	}

	if len(postsInRangeLastPage) == 0 {
		lastPage := findPageInRange(tid, firstPage.Extra.TotalPage, to, true)
		if postsInRangeLastPage, noMore = ExtractPostInRange(lastPage.PostList, from, to, true); noMore {
			tx.PostList = postsInRangeLastPage
			return tx
		}
	}

	tx.PostList = ScanPost(firstPage.Extra.CurrentPage, lastPage.Extra.CurrentPage-1, postsInRangeLastPage)

	return tx
}

//这里的thread不是主题,是线程..
const THREAD_NUMBER = 500

func ScanPost(fromPn, toPn int, postList []thread.ThreadPagePost) []thread.ThreadPagePost {

}

func findPageInRange(tid uint64, toPn int, toTime time.Time, findLastPage bool) *Thread {

	var fromPn = 1
	var _toPn = toPn
	var lastPn int
	var page *Thread

	if findLastPage {
		fmt.Print("寻找在分析范围内的最后一贴所在的页面:")
	} else {
		fmt.Print("寻找在分析范围内的最前一贴所在的页面:")
	}
	defer func() { fmt.Print("\n") }()

	for pn := (fromPn + toPn) / 2; ; pn = (fromPn + toPn) / 2 {
		if pn == 0 {
			pn = 1
		}
		if pn == lastPn {
			if findLastPage {
				if page.PostList[0].PostTime.After(toTime) {
					pn--
					if pn == 0 {
						fmt.Print("不存在.")
						return nil
					}
					page, _ = tryGettingThreadStruct(tid, pn)
				}
				fmt.Print(pn, "找到.")
				return page
			} else {
				if page.PostList[len(page.PostList)-1].PostTime.Before(toTime) {
					pn++
					if pn == _toPn+1 {
						fmt.Print("不存在.")
						return nil
					}
					page, _ = tryGettingThreadStruct(tid, pn)
				}
				fmt.Print(pn, "找到.")
				return page
			}

		}
		page, _ = tryGettingThreadStruct(tid, pn)

		if page.PostList[len(page.PostList)-1].PostTime.Before(toTime) {
			fmt.Print(pn, "过前->")
			fromPn = pn + 1
		} else if page.PostList[0].PostTime.After(toTime) {
			fmt.Print(pn, "过后->")
			toPn = pn - 1
		} else {
			fmt.Print(pn, "找到.")
			return page
		}
		lastPn = pn
	}

}

func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func ExtractPostInRange(postList []thread.ThreadPagePost, from, to time.Time, isLastPage bool) (postsInRage []thread.ThreadPagePost, noMore bool) {
	var x, y int

	if postList[0].PostTime.After(to) || postList[len(postList)-1].PostTime.Before(from) {
		return nil, false
	}
	//panic(1)
	for x = 0; x < len(postList); x++ {
		if postList[x].PostTime.Unix() >= from.Unix() {
			break
		}
	}
	for y = x; y < len(postList); y++ {
		if postList[y].PostTime.Unix() > to.Unix() {
			break
		}
	}

	if isLastPage {
		if x != 0 {
			noMore = true
		}
	} else {
		if y != len(postList) {
			noMore = true
		}
	}

	postsInRage = postList[x:y]
	return
}

type Thread struct {
	Page     *thread.ThreadPage
	PostList []thread.ThreadPagePost
	Extra    *thread.ThreadPageExtra
}

func tryGettingThreadStruct(tid uint64, pn int) (*Thread, *pberrors.PbError) {
	for {
		page, postList, extra, err, pberr := thread.GetThreadStruct(acc, tid, false, 0, pn, 30, false, true, false)
		if err == nil {
			return &Thread{page, postList, extra}, pberr
		}
	}

}
