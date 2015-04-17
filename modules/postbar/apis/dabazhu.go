package apis

import (
	"encoding/json"

	"github.com/purstal/pbtools/modules/http"
	"github.com/purstal/pbtools/modules/pberrors"
	"github.com/purstal/pbtools/modules/postbar/accounts"
)

func CancelBlockIDWeb(account *accounts.Account,
	forumName, userID,
	userName string) (error, *pberrors.PbError) {

	var parameters http.Parameters
	parameters.Add("word", forumName)

	tbs, err, pberr := account.GetTbs()
	if err != nil {
		return err, nil
	} else if pberr != nil {
		return nil, pberr
	}
	parameters.Add("tbs", tbs)

	parameters.Add("ie", "gbk")
	parameters.Add("type", "0")
	parameters.Add("list%5B0%5D%5Buser_id%5D", userID)
	parameters.Add("list%5B0%5D%5Buser_name%5D", userName)

	var cookies http.Cookies
	cookies.Add("BDUSS", account.BDUSS)

	resp, err := http.Get("http://tieba.baidu.com/bawu2/platform/cancelFilter", parameters, cookies)

	if err != nil {
		return err, nil
	}
	var x struct {
		ErrorCode int    `json:"errno"`
		ErrorMsg  string `json:"errmsg"`
	}

	json.Unmarshal(resp, &x)

	return nil, pberrors.NewPbError(x.ErrorCode, x.ErrorMsg)

}
