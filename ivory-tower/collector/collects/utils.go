package collects

import (
	"fmt"

	"github.com/purstal/go-tieba-base/logs"
	"github.com/purstal/go-tieba-base/tieba"
	"github.com/purstal/go-tieba-base/tieba/apis/forum-win8-1.5.0.0"
)

func TryGettingForumPageThreads(accWin8 *postbar.Account, forumName string, rn, pn int) []*forum.ForumPageThread {
	var retryTimes int
	const MAX_RETRY_TIME = 2
	for {
		_, posts, _, err, pberr := forum.GetForumStruct(accWin8, forumName, rn, pn)
		if err == nil {
			if pberr == nil || pberr.ErrorCode == 0 {
				var realRn = rnOf(posts, pn)
				if realRn != rn {
					if rn < realRn {
						logs.Warn(fmt.Sprintf("本页所求rn:%d, 实际rn:%d; 本页pn:%d", rn, realRn, pn))
						return posts
					}
					if retryTimes < MAX_RETRY_TIME {
						retryTimes++
					} else {
						logs.Warn(fmt.Sprintf("经%d次尝试, 本页所求rn:%d, 实际rn:%d; 本页pn:%d", MAX_RETRY_TIME+1, rn, realRn, pn))
						return posts
					}
				} else {
					return posts
				}
			} else {
				logs.Error(fmt.Sprintf("无法获取主页,重试. pn:%d; pberror:%d(%s).", pn, pberr.ErrorCode, pberr.ErrorMsg))
			}
		} else {
			logs.Error(fmt.Sprintf("无法获取主页,重试. pn:%d; error:%s", pn, err.Error()))
		}
	}
}

func rnOf(threads []*forum.ForumPageThread, pn int) int {
	var rn = len(threads)
	if rn == 0 {
		return 0
	}
	if threads[0].IsLivePost && pn != 1 {
		rn--
	}
	return rn

}
