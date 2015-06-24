package postbar

import (
	"time"
)

type IAuthor interface {
	AGetID() (bool, uint64)
	AGetName() string
	AGetIsLike() (bool, bool)
	AGetLevel() (bool, uint8)
	AGetPortrait() (bool, string)
}

type PostType int

const (
	Post PostType = iota
	SubPost
)

type IPost interface {
	PGetPid() uint64
	PGetFloor() (bool, int)
	PGetPostTime() (bool, time.Time)
	PGetOriginalContentList() (bool, []interface{})
	PGetContentList() []Content
	PContentIsComplete() bool
	PGetAuthor() IAuthor
	//PGetThread() IThread
}

//type IComment interface {
//IPost
//CGetPost() IPost
//}

type IThread interface {
	TGetTid() uint64
	TGetTitle() string
	TGetReplyNum() (bool, uint32)
	TGetLastReplyTime() (bool, time.Time)
	TGetIsTop() (bool, bool)
	TGetIsGood() (bool, bool)
	TGetAuthor() IAuthor
	TGetLastReplyer() IAuthor
	//TGetOriginalContent() (bool, []interface{})
	TGetContentList() []Content
	TContentIsComplete() bool
	//TGetForum() IForum
}

type IForum interface {
	FGetFid() uint32
	FGetForumName() string
}

type IExtra interface {
	EGetServerTime() time.Time
}
