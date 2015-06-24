package forum

import (
	"strconv"
	"time"

	"github.com/purstal/pbtools/modules/logs"
	"github.com/purstal/pbtools/modules/postbar"
)

type ForumPageThreadAuthor struct {
	postbar.IAuthor
	Name     string
	ID       uint64
	Portrait string
}

func (a ForumPageThreadAuthor) AGetName() string             { return a.Name }
func (a ForumPageThreadAuthor) AGetID() (bool, uint64)       { return true, a.ID }
func (a ForumPageThreadAuthor) AGetIsLike() (bool, bool)     { return false, false }
func (a ForumPageThreadAuthor) AGetLevel() (bool, uint8)     { return false, 0 }
func (a ForumPageThreadAuthor) AGetPortrait() (bool, string) { return true, a.Portrait }

type ForumPageThreadRelpyer struct {
	postbar.IAuthor
	Name string
	ID   uint64
}

func (a ForumPageThreadRelpyer) AGetName() string             { return a.Name }
func (a ForumPageThreadRelpyer) AGetID() (bool, uint64)       { return true, a.ID }
func (a ForumPageThreadRelpyer) AGetIsLike() (bool, bool)     { return false, false }
func (a ForumPageThreadRelpyer) AGetLevel() (bool, uint8)     { return false, 0 }
func (a ForumPageThreadRelpyer) AGetPortrait() (bool, string) { return false, "" }

type ForumPageThread struct {
	postbar.IThread
	Tid      uint64
	Title    string
	ReplyNum uint32

	LastReplyTime time.Time

	IsTop, IsGood bool

	Author      ForumPageThreadAuthor
	LastReplyer ForumPageThreadRelpyer

	MediaList []interface{}
	Abstract  []interface{}

	//Forum *ForumPage
}

func (t ForumPageThread) TGetTid() uint64                      { return t.Tid }
func (t ForumPageThread) TGetTitle() string                    { return t.Title }
func (t ForumPageThread) TGetReplyNum() (bool, uint32)         { return true, t.ReplyNum }
func (t ForumPageThread) TGetLastReplyTime() (bool, time.Time) { return true, t.LastReplyTime }
func (t ForumPageThread) TGetIsTop() (bool, bool)              { return true, t.IsTop }
func (t ForumPageThread) TGetIsGood() (bool, bool)             { return true, t.IsGood }
func (t ForumPageThread) TGetAuthor() postbar.IAuthor          { return t.Author }
func (t ForumPageThread) TGetLastReplyer() postbar.IAuthor     { return t.LastReplyer }
func (t ForumPageThread) TGetOriginalContentList() (bool, []interface{}) {
	return true, append(t.Abstract, t.MediaList...)
}
func (t ForumPageThread) TGetContentList() []postbar.Content {

	var originalContentList = append(t.Abstract, t.MediaList...)
	var contents = make([]postbar.Content, 0)

	for _, originalContent := range originalContentList {
		var contentMap map[string]interface{}
		var ok bool
		if contentMap, ok = originalContent.(map[string]interface{}); !ok {
			logs.Error("获取内容中的一项失败", originalContent)
			contents = append(contents, originalContent)
			continue
		}

		var content postbar.Content

		func() {
			defer func() {
				err := recover()
				if err != nil {
					logs.Error("获取内容中的一项的属性失败:", err, contentMap)
					content = contentMap
				}
			}()

			var contentType int32
			switch contentMap["type"].(type) {
			case (string):
				contentType_str, _ := strconv.Atoi(contentMap["type"].(string))
				contentType = int32(contentType_str)
			case (float64):
				contentType = int32((contentMap["type"].(float64)))
			}
			switch contentType {
			case 0:
				content = postbar.Text{contentMap["text"].(string)}
			case 3:
				content = postbar.Pic{contentMap["big_pic"].(string)}
			case 5:
				content = postbar.Video{contentMap["vhsrc"].(string)}
			case 6:
				content = postbar.Music{contentMap["vhsrc"].(string)}
			default:
				content = contentMap
			}

			contents = append(contents, content)
		}()
	}
	return contents

}
func (t ForumPageThread) TContentIsComplete() bool { return false }

//func (t ForumPageThread) TGetForum() postbar.IForum    { return t.Forum }

type ForumPage struct {
	postbar.IForum

	Fid       uint32
	ForumName string
}

func (f ForumPage) FGetFid() uint32       { return f.Fid }
func (f ForumPage) FGetForumName() string { return f.ForumName }

type ForumPageExtra struct {
	postbar.IExtra
	IsLogin    bool
	ServerTime time.Time
	LogID      uint64
}

func (e ForumPageExtra) EGetServerTime() time.Time {
	return e.ServerTime
}
