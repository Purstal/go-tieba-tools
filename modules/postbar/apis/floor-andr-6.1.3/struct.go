package floor

import (
	"time"

	"github.com/purstal/pbtools/modules/postbar"
	"github.com/purstal/pbtools/modules/postbar/apis/thread-win8-1.5.0.0"
)

type FloorPageComment struct {
	//interfaces.IComment
	postbar.IPost
	Spid        uint64
	ContentList []interface{}
	PostTime    time.Time
	Author      FloorPageCommentAuthor
	//Thread      threadpage.ThreadPage
	//Post        threadpage.ThreadPagePost
}

func (c FloorPageComment) PGetPid() uint64                                { return c.Spid }
func (c FloorPageComment) PGetFloor() (bool, int)                         { return false, 0 }
func (c FloorPageComment) PGetPostTime() (bool, time.Time)                { return true, c.PostTime }
func (c FloorPageComment) PGetOriginalContentList() (bool, []interface{}) { return true, c.ContentList }
func (c FloorPageComment) PGetContentList() []postbar.Content {
	return thread.ExtractContent(c.ContentList)
}                                                      /*...*/
func (c FloorPageComment) PContentIsComplete() bool    { return true }
func (c FloorPageComment) PGetAuthor() postbar.IAuthor { return c.Author }

type FloorPageCommentAuthor struct {
	postbar.IAuthor
	ID       uint64
	Name     string
	Level    uint8
	Portrait string
}

func (a FloorPageCommentAuthor) AGetID() (bool, uint64)       { return true, a.ID }
func (a FloorPageCommentAuthor) AGetName() string             { return a.Name }
func (a FloorPageCommentAuthor) AGetLevel() (bool, uint8)     { return true, a.Level }
func (a FloorPageCommentAuthor) AGetIsLike() (bool, bool)     { return false, false }
func (a FloorPageCommentAuthor) AGetPortrait() (bool, string) { return true, a.Portrait }

type FloorPageExtra struct {
	postbar.IExtra
	ServerTime time.Time
}

func (e FloorPageExtra) EGetServerTime() time.Time {
	return e.ServerTime
}
