package thread

import (
	"math"
	"time"

	"github.com/purstal/pbtools/modules/postbar"
	"github.com/purstal/pbtools/modules/postbar/forum-win8-1.5.0.0"
)

type ThreadPageAuthorAndThreadPagePostAuthor struct {
	postbar.IAuthor
	ID        uint64
	Name      string
	HasIsLike bool
	IsLike    bool
	Level     uint8
	Portrait  string
}

func (a ThreadPageAuthorAndThreadPagePostAuthor) AGetID() (bool, uint64) { return true, a.ID }
func (a ThreadPageAuthorAndThreadPagePostAuthor) AGetName() string       { return a.Name }
func (a ThreadPageAuthorAndThreadPagePostAuthor) AGetIsLike() (bool, bool) {
	return a.HasIsLike, a.IsLike
}
func (a ThreadPageAuthorAndThreadPagePostAuthor) AGetLevel() (bool, uint8) {
	return a.Level != math.MaxUint8, a.Level
}
func (a ThreadPageAuthorAndThreadPagePostAuthor) AGetPortrait() (bool, string) {
	return true, a.Portrait
}

type ThreadPagePost struct {
	postbar.IPost
	Pid   uint64
	Floor int

	PostTime time.Time

	ContentList []interface{}

	Author ThreadPageAuthorAndThreadPagePostAuthor

	//Thread *ThreadPage
}

func (p ThreadPagePost) PGetPid() uint64                                { return p.Pid }
func (p ThreadPagePost) PGetFloor() (bool, int)                         { return true, p.Floor }
func (p ThreadPagePost) PGetPostTime() (bool, time.Time)                { return true, p.PostTime }
func (p ThreadPagePost) PGetOriginalContentList() (bool, []interface{}) { return true, p.ContentList }
func (p ThreadPagePost) PGetContentList() []postbar.Content             { return ExtractContent(p.ContentList) } /*...*/
func (p ThreadPagePost) PContentIsComplete() bool                       { return true }
func (p ThreadPagePost) PGetAuthor() postbar.IAuthor                    { return p.Author }

//func (p ThreadPagePost) PGetThread() interfaces.IThread                 { return p.Thread }

type ThreadPage struct {
	ForumPageThread *forum.ForumPageThread
	//HasForumPageThread bool
	postbar.IThread
	Tid    uint64
	Title  string
	Author ThreadPageAuthorAndThreadPagePostAuthor
}

func (t ThreadPage) TGetTid() uint64   { return t.Tid }
func (t ThreadPage) TGetTitle() string { return t.Title }
func (t ThreadPage) TGetReplyNum() (bool, uint32) {
	if t.ForumPageThread != nil {
		return t.ForumPageThread.TGetReplyNum()
	}
	return false, 0
}
func (t ThreadPage) TGetLastReplyTime() (bool, time.Time) {
	if t.ForumPageThread != nil {
		return t.ForumPageThread.TGetLastReplyTime()
	}
	return false, time.Time{}

}
func (t ThreadPage) TGetIsTop() (bool, bool) {

	if t.ForumPageThread != nil {
		return t.ForumPageThread.TGetIsTop()
	}
	return false, false
}
func (t ThreadPage) TGetIsGood() (bool, bool) {
	if t.ForumPageThread != nil {
		return t.ForumPageThread.TGetIsGood()
	}
	return false, false
}
func (t ThreadPage) TGetAuthor() postbar.IAuthor { return t.Author }
func (t ThreadPage) TGetLastReplyer() (bool, postbar.IAuthor) {
	if t.ForumPageThread != nil {
		return t.ForumPageThread.TGetLastReplyer()
	}
	return false, forum.ForumPageThreadAuthor{}

}
func (t ThreadPage) TGetOriginalContentList() (bool, []interface{}) {
	if t.ForumPageThread != nil {
		return t.ForumPageThread.TGetOriginalContentList()
	}
	return false, nil
}
func (t ThreadPage) TGetContentList() []postbar.Content {
	if t.ForumPageThread != nil {
		return t.ForumPageThread.TGetContentList()
	}
	return nil
}
func (t ThreadPage) TContentIsComplete() bool {
	if t.ForumPageThread != nil {
		return t.ForumPageThread.TContentIsComplete()
	}
	return false

}

//func (t ThreadPage) TGetForum() interfaces.IForum    { return t.ForumPageThread.TGetForum() }

//func (t ThreadPage) TGetAuthor() interfaces.IAuthor { return t.Author }

type ThreadPageExtra struct {
	postbar.IExtra
	CurrentPage int
	TotalPage   int
	ServerTime  time.Time
}

func (e ThreadPageExtra) EGetServerTime() time.Time {
	return e.ServerTime
}
