package forum_page_monitor

import (
	"github.com/purstal/pbtools/modules/postbar/forum-win8-1.5.0.0"
)

type PageThreadsSorter []*forum.ForumPageThread

func (sorter PageThreadsSorter) Len() int {
	return len(sorter)

}

func (sorter PageThreadsSorter) Less(i, j int) bool {
	return sorter[i].LastReplyTime.After(sorter[j].LastReplyTime)
}

func (sorter PageThreadsSorter) Swap(i, j int) {
	sorter[i], sorter[j] = sorter[j], sorter[i]
}
