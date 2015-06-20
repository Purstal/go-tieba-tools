package post_deleter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/purstal/pbtools/modules/postbar"
	tp "github.com/purstal/pbtools/modules/postbar/apis/thread-win8-1.5.0.0"

	postfinder "github.com/purstal/pbtools/tool-core/post-finder"
)

func (d *PostDeleter) NewThreadFirstAssessor(account *postbar.Account, thread *postfinder.ForumPageThread) postfinder.Control {
	keyWords := d.Content_RxKw.KeyWords()
	var deleteReason string
	var banFlag bool
	if matchedExp := MatchAny(thread.Thread.Title, keyWords); matchedExp != nil {
		deleteReason = fmt.Sprint("标题匹配关键词:", matchedExp)
		banFlag = matchedExp.BanFlag
	} else if matchedExp := MatchAny(ExtractText(thread.Thread.TGetContentList()), keyWords); matchedExp != nil {
		deleteReason = fmt.Sprint("内容匹配关键词:", matchedExp)
		banFlag = matchedExp.BanFlag
	} else if strings.Contains(thread.Thread.Title, "乡村") &&
		!strings.Contains(thread.Thread.Title, "改造") && !strings.Contains(thread.Thread.Title, "建筑") {
		deleteReason = "乡村类垃圾主题"
	} else if match, _ := regexp.MatchString(`(传奇.*?[A-Za-z0-9]{2}|[0-9A-Za-z]{2}.*?传奇)`, thread.Thread.Title); match {
		deleteReason = "传奇私服广告"
	}

	if deleteReason != "" {
		d.DeleteThread("主页页面", account, thread.Thread.Tid, 0, thread.Thread.Author.ID, deleteReason)
		if banFlag {
			pid := GetPidFromTid(thread.Thread.Tid, d.AccWin8)
			if pid == 0 {
				d.Logger.Error(MakePrefix(nil, thread.Thread.Tid, pid, 0, thread.Thread.Author.ID),
					"无法获取主题pid,无法进行封禁,将不进行封禁.")
			} else {
				d.BanID("主页页面", account.BDUSS, thread.Thread.Author.Name,
					d.ForumID, thread.Thread.Tid, pid, thread.Thread.Author.ID, 1, deleteReason, "null")

			}
		}
		return postfinder.Finish
	}

	return postfinder.Continue
}

func (d *PostDeleter) NewThreadSecondAssessor(account *postbar.Account, post *postfinder.ThreadPagePost) {
	if d.CommonAssess("主题页面(新主题)", account, post.Post, post.Thread.Tid) == postfinder.Finish {
		return
	}
}

func GetPidFromTid(tid uint64, accWin8 *postbar.Account) uint64 {
	for i := 0; ; {
		thread, err, pberr := tp.GetThread2(accWin8, tid, false, 0, 1, 2, false, true, false)
		if err != nil {
			continue
		}
		if pberr != nil {
			if pberr.ErrorCode == 4 || i >= 3 {
				return 0
			}
			i++
		} else {
			if len(thread.PostList) == 0 {
				return 0
			}
			return thread.PostList[0].Pid
		}
	}
}
