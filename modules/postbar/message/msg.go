package message

import (
	"encoding/json"

	"github.com/purstal/pbtools/modules/http"
	"github.com/purstal/pbtools/modules/pberrors"
	"github.com/purstal/pbtools/modules/postbar"
)

/*
{
	"message":{
		"fans":"0",
		"replyme":"2",
		"atme":"0",
		"pletter":"212",
		"bookmark":"0",
		"count":"214"
		},
	"server_time":"132295",
	"time":1432897589,
	"ctime":0,
	"logid":389046366,
	"error_code":"0"
}
*/

type Notice struct {
	Fans     int `json:"fans,string"`
	ReplyMe  int `json:"replyme,string"`
	AtMe     int `json:"atme,string"`
	PLetter  int `json:"pletter,string"`
	BookMark int `json:"bookmark,string"`
	Count    int `json:"count,string"`
}

func GetNotice(acc *postbar.Account) (*Notice, error, *pberrors.PbError) {
	var parameters http.Parameters
	postbar.ProcessParams(&parameters, acc)
	resp, err := http.Post("http://c.tieba.baidu.com"+"/c/s/msg", parameters)
	if err != nil {
		return nil, err, nil
	}
	var message struct {
		Notice    Notice `json:"message"`
		ErrorCode int    `json:"error_code,string"`
		ErrorMsg  string `json:"error_msg"`
	}

	if message.ErrorCode != 0 {
		return nil, nil, pberrors.NewPbError(message.ErrorCode, message.ErrorMsg)
	}

	err = json.Unmarshal(resp, &message)
	return &message.Notice, err, nil
}
