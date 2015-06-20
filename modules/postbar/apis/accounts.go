package apis

import (
	"encoding/json"

	"github.com/purstal/pbtools/modules/http"
	"github.com/purstal/pbtools/modules/misc"
	"github.com/purstal/pbtools/modules/postbar"
)

func RGetTbs(acc *postbar.Account) ([]byte, error) {
	var parameters http.Parameters
	postbar.ProcessParams(&parameters, acc)
	return http.Post(`http://c.tieba.baidu.com/c/s/tbs`, parameters)
}

func RGetTbsWeb(BDUSS string) ([]byte, error) {
	var cookies http.Cookies
	cookies.Add("BDUSS", BDUSS)
	return http.Get(`http://tieba.baidu.com/dc/common/tbs`, nil, cookies)

}

func RLogin(acc *postbar.Account, username, password string) ([]byte, error) {

	var parameters http.Parameters
	parameters.Add("un", username)
	parameters.Add("passwd", misc.ComputeBase64(password))
	postbar.ProcessParams(&parameters, acc)
	return http.Post("http://c.tieba.baidu.com/c/s/login", parameters)
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

func Login(acc *postbar.Account, password string) (error, *postbar.PbError) {
	//resp, err := APILogin(acc, acc.ID, password)
	resp, err := RLogin(acc, acc.ID, password)
	if err != nil {
		return err, nil
	}
	var x struct {
		ErrorCode int    `json:"error_code,string"`
		ErrorMsg  string `json:"error_msg"`
		User      struct {
			BDUSS string `json:"BDUSS"`
		} `json:"user"`
	}

	err2 := json.Unmarshal(resp, &x)

	if err2 != nil {
		return err2, nil
	}
	if x.ErrorCode != 0 {
		return nil, postbar.NewPbError(x.ErrorCode, x.ErrorMsg)
	}
	acc.BDUSS = x.User.BDUSS
	return nil, nil
}

func IsLogin(BDUSS string) (bool, error) {
	resp, err := RGetTbsWeb(BDUSS)
	if err != nil {
		return false, err
	}
	var x struct {
		Tbs     string `json:"tbs"`
		IsLogin int    `json:"is_login"`
	}
	err2 := json.Unmarshal(resp, &x)
	if err2 != nil {
		return false, err2
	}
	return x.IsLogin == 1, nil
}
