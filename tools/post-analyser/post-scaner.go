package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/purstal/go-tieba-base/http"
	"github.com/purstal/go-tieba-base/tieba"
	"github.com/purstal/go-tieba-base/tieba/thread-win8-1.5.0.0"
	"github.com/purstal/pbtools/tool-cores/utils"
)

var acc = postbar.NewDefaultWindows8Account("")

func main() {
	http.RetryTimes = 1
	http.ShutUp = true

	usage := fmt.Sprintf(`usage:
%s $tid $from-time $to-time

$from-time $to-time: 如 2006-01-02`, os.Args[0])
	if len(os.Args) != 4 {
		fmt.Println(usage)
		return
	}

	var tid uint64
	var err error
	var fromTime, toTime time.Time

	if tid, err = strconv.ParseUint(os.Args[1], 10, 64); err != nil {
		fmt.Println("tid格式不对,", err, ".输入:", os.Args[1])
		fmt.Println(usage)
		return
	}
	if fromTime, err = parseTime(os.Args[2]); err != nil {
		fmt.Println("fromTime格式不对,", err, ".输入:", os.Args[2])
		fmt.Println(usage)
		return
	}
	if toTime, err = parseTime(os.Args[3]); err != nil {
		fmt.Println("toTime格式不对,", err, ".输入:", os.Args[3])
		fmt.Println(usage)
		return
	}
	toTime = time.Unix(toTime.Unix()+24*60*60-1, 0)

	//fromTime := time.Date(2012, 8, 20, 23, 57, 0, 0, time.Local)
	//2012-08-20 23:57
	//toTime := time.Date(2013, 4, 2, 9, 23, 0, 0, time.Local)
	//2012-10-11 02:19

	tx := CollectPost(tid, fromTime, toTime, 10)

	var dir = "scanned-thread/"
	os.MkdirAll(dir, 0644)
	f, err := os.Create(dir + strconv.FormatUint(tid, 10) + time.Now().Format("[20060102-150304].json"))
	if err != nil {
		fmt.Println("无法创造文件,将不保存:", err, ".")
	} else {
		j, err := json.Marshal(tx)
		if err != nil {
			fmt.Println("编排json文件失败,将不保存:", err, ".")
		} else {
			f.Write(j)
		}
	}
}

func parseTime(str string) (time.Time, error) {
	var year, month, day int
	_, err := fmt.Sscanf(str, "%d-%d-%d",
		&year, &month, &day)
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local), err
}

type ThreadX struct {
	Version string
	Info    struct {
		Author string
		Title  string
	}
	ScanInfo struct {
		FromTime time.Time
		ToTime   time.Time
		ScanTime time.Time
	}
	PostList []thread.ThreadPagePost
}

func CollectPost(tid uint64, from, to time.Time, maxScanThreadNumber int) ThreadX {
	var tx ThreadX
	tx.Version = "1"
	tx.ScanInfo.FromTime = from
	tx.ScanInfo.ToTime = to
	tx.ScanInfo.ScanTime = time.Now()

	firstPage, pberr := tryGettingThreadStruct(tid, 1)

	tx.Info.Author = firstPage.Page.Author.Name
	tx.Info.Title = firstPage.Page.Title

	if pberr != nil && pberr.ErrorCode != 0 {
		fmt.Println("无法获取主题", tid, "第一页:", pberr, ",放弃.")
		return tx
	}

	//var postsInRange []thread.ThreadPagePost
	var /*postsInRangeFirstPage,*/ postsInRangeLastPage []thread.ThreadPagePost
	var noMore bool

	if tx.PostList, noMore = ExtractPostInRange(firstPage.PostList, &from, &to, false); noMore {
		return tx
	}

	lastPage, _ := tryGettingThreadStruct(tid, firstPage.Extra.TotalPage)
	if postsInRangeLastPage, noMore = ExtractPostInRange(lastPage.PostList, &from, &to, true); noMore {
		tx.PostList = postsInRangeLastPage
		return tx
	}

	if len(tx.PostList) == 0 {
		firstPage = findPageInRange(tid, firstPage.Extra.TotalPage, from, false)
		if tx.PostList, noMore = ExtractPostInRange(firstPage.PostList, &from, &to, false); noMore {
			return tx
		}
	}

	if len(postsInRangeLastPage) == 0 {
		lastPage = findPageInRange(tid, firstPage.Extra.TotalPage, to, true)
		if postsInRangeLastPage, noMore = ExtractPostInRange(lastPage.PostList, &from, &to, true); noMore {
			tx.PostList = postsInRangeLastPage
			return tx
		}
	}

	tx.PostList = ScanPost(tid, firstPage.Extra.CurrentPage,
		lastPage.Extra.CurrentPage-1, from, postsInRangeLastPage, maxScanThreadNumber)

	return tx
}

//这里的thread不是主题,是线程..

func ScanPost(tid uint64, fromPn, toPn int, fromTime time.Time, postListLastPage []thread.ThreadPagePost, maxThreadNumber int) []thread.ThreadPagePost {
	b := time.Now()
	fmt.Printf("正在扫描主题:%d.从第%d页至第%d页.\n", tid, fromPn, toPn)
	var realThreadNumber int
	if realThreadNumber = toPn - fromPn + 1; realThreadNumber > maxThreadNumber {
		realThreadNumber = maxThreadNumber
	}
	fmt.Println("使用线程数:", realThreadNumber, ".")

	var totalPn = toPn - fromPn + 1

	var postLists = make([][]thread.ThreadPagePost, totalPn)
	postLists[len(postLists)-1] = postListLastPage

	manager := utils.NewLimitTaskManager(realThreadNumber, totalPn)
	var lastLineCount int
	for i := 0; i < toPn-fromPn+1; i++ {
		manager.RequireChan <- true
		if <-manager.DoChan {
			go func(i int) {
				var format = "\r扫描进度:%06.2f%%(%d/%d(活跃线程数:%03d))."
				lastLineCount = len(format)
				fmt.Printf(format,
					float64(manager.FinishedCount*100)/float64(totalPn),
					manager.FinishedCount, totalPn, manager.ActiveCount)
				postList, _ := tryGettingThreadStruct(tid, fromPn+i)
				postLists[i] = postList.PostList
				manager.FinishChan <- true
			}(i)
		} else {

		}
	}
	<-manager.AllTaskFinishedChan
	fmt.Println("\n完成扫描.耗时", time.Now().Sub(b).String(), ".")

	postList, _ := ExtractPostInRange(postLists[0], &fromTime, nil, false)

	for i := 1; i < len(postLists); i++ {
		for j := len(postList) - 1; j >= 0; j-- {
			if postList[j].Pid < postLists[i][0].Pid {
				postList = append(postList[:j+1], postLists[i]...)
				break
			}
			//???
		}
	}
	return postList
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

func ExtractPostInRange(postList []thread.ThreadPagePost, from, to *time.Time, isLastPage bool) (postsInRage []thread.ThreadPagePost, noMore bool) {
	var x, y int

	if (to != nil && postList[0].PostTime.After(*to)) ||
		(from != nil && postList[len(postList)-1].PostTime.Before(*from)) {
		return nil, false
	}
	//panic(1)
	if from != nil {
		for x = 0; x < len(postList); x++ {
			if postList[x].PostTime.Unix() >= from.Unix() {
				break
			}
		}
	} else {
		x = 0
	}

	if to != nil {
		for y = x; y < len(postList); y++ {
			if postList[y].PostTime.Unix() > to.Unix() {
				break
			}
		}
	} else {
		y = len(postList)
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
	//fmt.Println(x, y)
	postsInRage = postList[x:y]
	return
}

type Thread struct {
	Page     *thread.ThreadPage
	PostList []thread.ThreadPagePost
	Extra    *thread.ThreadPageExtra
}

func tryGettingThreadStruct(tid uint64, pn int) (*Thread, *postbar.PbError) {
	for {
		page, postList, extra, err, pberr := thread.GetThreadStruct(acc, tid, false, 0, pn, 30, false, true, false)
		if err == nil {
			return &Thread{page, postList, extra}, pberr
		}
		fmt.Printf("尝试获取主题页pn=%d失败,半秒钟后重试:%s.\n", pn, err)
		time.Sleep(time.Second / 2)
	}

}
