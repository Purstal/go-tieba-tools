package apis

import (
	"encoding/json"

	"github.com/purstal/pbtools/modules/http"
	"github.com/purstal/pbtools/modules/misc"
	"github.com/purstal/pbtools/modules/pberrors"
)

func GetFid(fname string) (uint64, error, *pberrors.PbError) {
	var parameters http.Parameters
	parameters.Add("fname", misc.UrlQueryEscape(misc.ToGBK(fname)))
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
		return 0, nil, pberrors.NewPbError(x.ErrorNo, x.ErrorMsg)
	}

	json.Unmarshal(resp, &x)

	return x.Data.Fid, nil, nil

}

func GetUid(un string) (uint64, error) { //忠于百度的写法,用Get取
	var parameters http.Parameters
	parameters.Add("un", misc.UrlQueryEscape(misc.ToGBK(un)))
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

func HasWrongUserJson(un string) (bool, error) {
	var parameters http.Parameters
	parameters.Add("un", misc.UrlQueryEscape(misc.ToGBK(un)))
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

func GetTbsWeb(BDUSS string) (string, error) {
	var cookies http.Cookies
	cookies.Add("BDUSS", BDUSS)

	data, err := http.Get(`http://tieba.baidu.com/dc/common/tbs`, nil, cookies)

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
