package apis

import (
	"github.com/purstal/pbtools/modules/http"
	"github.com/purstal/pbtools/modules/misc"
	"github.com/purstal/pbtools/modules/postbar"
)

func RTest(acc *postbar.Account, userName, password string) (
	[]byte, error) {

	var parameters http.Parameters
	parameters.Add("un", userName)
	parameters.Add("passwd", misc.ComputeBase64(password))
	postbar.ProcessParams(&parameters, acc)
	return http.Post("http://c.tieba.baidu.com/c/s/test", parameters)
}
