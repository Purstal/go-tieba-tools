package apis

import (
	"encoding/json"

	"github.com/purstal/pbtools/modules/http"
	"github.com/purstal/pbtools/modules/misc"
	"github.com/purstal/pbtools/modules/postbar"
)

func GetFid(forumName string) (uint64, error, *postbar.PbError) {
	var parameters http.Parameters
	parameters.Add("fname", misc.ToGBK(forumName))
	resp, err := http.Post(`http://tieba.baidu.com/f/commit/share/fnameShareApi`, parameters)

	if err != nil {
		return 0, err, nil
	}

	var x struct {
		ErrorNo  int    `json:"no"`
		ErrorMsg string `json:"error"`
		Data     struct {
			Fid uint64 `json:"fid"`
		} `json:"data"`
	}

	if x.ErrorNo != 0 {
		return 0, nil, postbar.NewPbError(x.ErrorNo, x.ErrorMsg)
	}

	json.Unmarshal(resp, &x)

	return x.Data.Fid, nil, nil

}

func GetUid(userName string) (uint64, error) { //忠于百度的写法,用Get取
	var parameters http.Parameters
	parameters.Add("un", misc.UrlQueryEscape(misc.ToGBK(userName)))
	resp, err := http.Get("http://tieba.baidu.com/i/sys/user_json", parameters, nil)
	//println(string(resp))
	if err != nil {
		return 0, err
	}

	var x struct {
		Creator struct {
			ID uint64 `json:"id"`
		} `json:"creator"`
	}

	err2 := json.Unmarshal(resp, &x)

	if err2 != nil {
		return 0, err2
	}

	return x.Creator.ID, nil

}

func HasWrongUserJson(userName string) (bool, error) {
	var parameters http.Parameters
	parameters.Add("un", misc.UrlQueryEscape(misc.ToGBK(userName)))
	resp, err := http.Get("http://tieba.baidu.com/i/sys/user_json", parameters, nil)
	if err != nil {
		return false, err
	}

	var x struct {
		RawName string `json:"raw_name"`
	}

	err2 := json.Unmarshal(resp, &x)

	if err2 != nil {
		return false, err2
	}

	return x.RawName == "", nil
}

func RGetTbs(acc *postbar.Account) ([]byte, error) {
	var parameters http.Parameters
	postbar.ProcessParams(&parameters, acc)
	return http.Post(`http://c.tieba.baidu.com/c/s/tbs`, parameters)
}

func GetTbs(acc *postbar.Account) (string, error, *postbar.PbError) {
	resp, err := RGetTbs(acc)
	if err != nil {
		return "", err, nil
	}
	var x struct {
		ErrorCode int    `json:"error_code,string"`
		ErrorMsg  string `json:"error_msg"`
		Tbs       string `json:"tbs"`
	}
	err2 := json.Unmarshal(resp, &x)
	if err2 != nil {
		return "", err, nil
	}
	if x.ErrorCode != 0 {
		return "", nil, postbar.NewPbError(x.ErrorCode, x.ErrorMsg)
	}
	return x.Tbs, nil, nil
}

func RGetTbsWeb(BDUSS string) ([]byte, error) {
	var cookies http.Cookies
	cookies.Add("BDUSS", BDUSS)
	return http.Get(`http://tieba.baidu.com/dc/common/tbs`, nil, cookies)

}

func GetTbsWeb(BDUSS string) (string, error) {

	data, err := GetTbsWeb(BDUSS)

	if err != nil {
		return "", err
	}

	var x struct {
		Tbs string `json:"tbs"`
	}

	err2 := json.Unmarshal(data, &x)

	if err2 != nil {
		return "", err2
	}

	return x.Tbs, nil

}
