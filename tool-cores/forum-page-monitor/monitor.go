package forum_page_monitor

import (
	"fmt"
	"time"

	"github.com/purstal/pbtools/modules/postbar"
	"github.com/purstal/pbtools/modules/postbar/apis/forum-win8-1.5.0.0"
)

type action struct {
	action string
	param  interface{}
}

type ForumPageMonitor struct {
	rn       int
	PageChan chan ForumPage
	actChan  chan action
}

type ForumPage struct {
	Forum      *forum.ForumPage
	ThreadList []*forum.ForumPageThread
	Extra      *forum.ForumPageExtra
}

func (monitor *ForumPageMonitor) ChangeAccount(accWin8 *postbar.Account) {
	monitor.actChan <- action{"ChangeAccount", accWin8}
}

func (monitor *ForumPageMonitor) ChangeInterval(newInterval time.Duration) {
	monitor.actChan <- action{"ChangeInterval", newInterval}
}

func (monitor *ForumPageMonitor) Stop() {
	monitor.actChan <- action{"Stop", nil}
}

func NewForumPageMonitor(accWin8 *postbar.Account, kw string,
	interval time.Duration) *ForumPageMonitor {

	var monitor = ForumPageMonitor{}

	monitor.actChan = make(chan action)

	monitor.rn = 30

	var ticker = time.NewTicker(interval)
	monitor.PageChan = make(chan ForumPage)
	go func() {
		for {
			go func() {
				fp, fpts, fpe, err, pberr := forum.GetForumStruct(accWin8, kw, monitor.rn, 1)
				if err != nil || pberr != nil {
					//fmt.Println(accWin8)
					fmt.Println("获取主页时出错: ", err, pberr)
					return
				}
				monitor.PageChan <- ForumPage{
					Forum:      fp,
					ThreadList: fpts,
					Extra:      fpe,
				}
			}()
			select {
			case <-ticker.C:
			case act := <-monitor.actChan:
				switch act.action {
				case "ChangeAccount":
					accWin8 = act.param.(*postbar.Account)
				case "ChangeInterval":
					ticker.Stop()
					ticker = time.NewTicker(act.param.(time.Duration))
				case "Stop":
					close(monitor.actChan)
					close(monitor.PageChan)
					return
				}
			}
			//fmt.Println("本地时间:", time)
		}
	}()
	return &monitor
}
