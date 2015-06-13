package apis

import (
	"github.com/purstal/pbtools/modules/postbar"
)

type ForumList struct {
	CanUse  bool
	Content string //用于记录信息
	Level   int
}

func GetForumList(acc *postbar.Account) {

}
