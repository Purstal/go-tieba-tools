package accounts

import (
	"github.com/purstal/pbtools/modules/http"
	"github.com/purstal/pbtools/modules/misc"
)

func GetTbs(账号 *Account) ([]byte, error) { //Get_tbs([*Account])
	var parameters http.Parameters
	ProcessParams(&parameters, 账号)
	return http.Post(`http://c.tieba.baidu.com/c/s/tbs`, parameters)
}

func GetTbsWeb(BDUSS string) ([]byte, error) {
	var cookies http.Cookies
	cookies.Add("BDUSS", BDUSS)
	return http.Get(`http://tieba.baidu.com/dc/common/tbs`, nil, cookies)

}

func Login(账号 *Account, username, password string) ([]byte, error) {

	var parameters http.Parameters
	parameters.Add("un", username)
	parameters.Add("passwd", misc.ComputeBase64(password))
	ProcessParams(&parameters, 账号)
	return http.Post("http://c.tieba.baidu.com/c/s/login", parameters)

}
