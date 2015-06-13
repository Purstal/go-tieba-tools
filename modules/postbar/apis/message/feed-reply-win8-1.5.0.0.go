package message

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/purstal/pbtools/modules/pberrors"
	"github.com/purstal/pbtools/modules/postbar"
	"github.com/purstal/pbtools/modules/postbar/apis"
)

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

type OriginalReplyMessageStruct struct {
	IsFloor string `json:"is_floor"`
	Type    int    `json:"type,string"`
	Replyer struct {
		ID       uint64 `json:"id,string"`
		Name     string
		Portrait string
		IsFriend string
	}
	QuoteUser struct {
		ID   uint64 `json:"id,string"`
		Name string
	}
	Title        string
	Content      string
	QuoteContent string
	Tid          uint64 `json:"thread_id,string"`
	Pid          uint64 `json:"post_id,string"`
	Time         int64  `json:"time,string"`
	ForumName    string `json:"fname"`
	QuotePid     string `json:"quote_pid="`
}

func GetOriginalReplyMessageStruct(acc *postbar.Account) ([]OriginalReplyMessageStruct, error, *pberrors.PbError) {

	resp, err := apis.RFeedReplyMe(acc)

	if err != nil {
		return nil, err, nil
	}

	var msg struct {
		ReplyList []OriginalReplyMessageStruct `json:"reply_list"`

		ErrorCode int    `json:"error_code,string"`
		ErrorMsg  string `json:"error_msg"`
	}

	err = json.Unmarshal(resp, &msg)

	if err != nil {
		return nil, err, nil
	}

	if msg.ErrorCode != 0 {
		return nil, nil, pberrors.NewPbError(msg.ErrorCode, msg.ErrorMsg)
	}

	return msg.ReplyList, nil, nil

}

func GettReplyMessage(acc *postbar.Account) ([]ReplyMessage, error, *pberrors.PbError) {

	_msgs, err, pberr := GetOriginalReplyMessageStruct(acc)

	if err != nil {
		return nil, err, nil
	}

	if pberr != nil && pberr.ErrorCode != 0 {
		return nil, nil, pberr
	}

	var msgs = make([]ReplyMessage, len(_msgs))

	for i, _msg := range _msgs {

		msg := &msgs[i]

		msg.IsFloor = _msg.IsFloor == "1"
		msg.Type = _msg.Type
		msg.Replyer.ID = _msg.Replyer.ID
		msg.Replyer.Name = _msg.Replyer.Name
		msg.Replyer.Portarait = _msg.Replyer.Portrait
		msg.Replyer.IsFriend = _msg.Replyer.IsFriend == "1"
		msg.QuoteUser.ID = _msg.QuoteUser.ID
		msg.QuoteUser.Name = _msg.QuoteUser.Name
		msg.Title = _msg.Title
		msg.Content = _msg.Content
		msg.QuoteContent = _msg.QuoteContent
		msg.Tid = _msg.Tid
		msg.Pid = _msg.Pid
		msg.Time = time.Unix(_msg.Time, 0)
		if _msg.QuotePid == "" {
			msg.QuotePid = 0
		} else {
			msg.QuotePid, _ = strconv.ParseUint(_msg.QuotePid, 10, 64)
		}

	}

	return msgs, nil, nil
}
