package forum

import (
	"encoding/json"
	//"os"

	"github.com/purstal/pbtools/modules/pberrors"
	"github.com/purstal/pbtools/modules/postbar"
	"github.com/purstal/pbtools/modules/postbar/apis"
)

type OriginalForumStruct struct {
	Forum struct {
		ID   uint32 `json:"id,string"`
		Name string `json:"name"`
	} `json:"forum"`

	Page struct {
		PageSize int `json:"page_size,string"`
	} `json:"page"`

	User struct {
		IsLogin int `json:"is_login,string"`
	} `json:"user"`

	ThreadList []struct {
		Tid         uint64 `json:"tid,string"` //为毛还有个完全一样的项"id"= =
		Title       string `json:"title"`
		ReplyNum    uint32 `json:"reply_num,string"`
		LastTimeInt int64  `json:"last_time_int,string"`

		IsTop  int `json:"is_top,string"`
		IsGood int `json:"is_good,string"`

		Author struct {
			ID       uint64 `json:"id,string"`
			Name     string `json:"name"`
			Portrait string `json:"portrait"`
		} `json:"author"`

		LastReplyer struct {
			ID   uint64 `json:"id,string"`
			Name string `json:"name"`
		} `json:"last_replyer"`

		Media    []interface{} `json:"media"`
		Abstract []interface{} `json:"abstract"`
		//Abstract []struct {
		//	Text string `json:"text"`
		//} `json:"abstract"`

	} `json:"thread_list"`

	Time  int64  `json:"time"`
	LogID uint64 `json:"logid"`

	ErrorCode int    `json:"error_code,string"`
	ErrorMsg  string `json:"error_msg"`
}

func GetOriginalForumStruct(
	acc *postbar.Account, kw string, rn,
	pn int) (*OriginalForumStruct,
	error, *pberrors.PbError) {

	resp, err := apis.RGetForum(acc, kw, rn, pn)

	//f, _ := os.Create("debug.json")
	//f.Write(resp)

	if err != nil {
		return nil, err, nil
	}
	var x OriginalForumStruct
	err2 := json.Unmarshal(resp, &x)
	if err2 != nil {
		return nil, err2, nil
	}
	if x.ErrorCode != 0 {
		return &x, nil, pberrors.NewPbError(x.ErrorCode, x.ErrorMsg)
	}
	return &x, nil, nil
}
