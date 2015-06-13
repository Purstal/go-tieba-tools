package post_finder

import (
	"github.com/purstal/pbtools/modules/postbar/apis/forum-win8-1.5.0.0"
)

func IsNewThread(thread *forum.ForumPageThread) bool {
	if thread.ReplyNum == 0 &&
		thread.Author.ID == thread.LastReplyer.ID {
		return true
	}
	return false
}

//func ExtractThreadLog(thread postbar.ForumPageThread) []string {
//	var slice = make([]string)
//
//}
