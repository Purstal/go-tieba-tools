package message

import (
	"time"

	"github.com/purstal/pbtools/modules/http"
	"github.com/purstal/pbtools/modules/postbar"
)

func RFeedReplyMe(acc *postbar.Account) ([]byte, error) {
	var parameters http.Parameters
	postbar.ProcessParams(&parameters, acc)
	return http.Post("http://c.tieba.baidu.com"+"/c/u/feed/replyme", parameters)
}

func RFeedAtMe(acc *postbar.Account) ([]byte, error) {
	var parameters http.Parameters
	postbar.ProcessParams(&parameters, acc)
	return http.Post("http://c.tieba.baidu.com"+"/c/u/feed/atme", parameters)
}

type ReplyMessage struct {
	IsFloor bool
	Type    int //?
	//Unread bool //不靠谱

	Replyer struct {
		ID        uint64 //uid
		Name      string
		Portarait string
		IsFriend  bool
	}
	QuoteUser struct {
		ID   uint64
		Name string
	}
	Title        string
	Content      string
	QuoteContent string
	Tid          uint64
	Pid          uint64
	Time         time.Time
	ForumName    string
	QuotePid     uint64
}
