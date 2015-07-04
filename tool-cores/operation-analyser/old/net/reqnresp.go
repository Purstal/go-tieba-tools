package net

import "net/http"
import "bytes"
import "io/ioutil"

import "github.com/purstal/pbtools/tool-cores/operation-analyser/old/log"

//http://www.crifan.com/go_language_http_do_post_pass_post_data/

func Get(srcUrl string, parameters Parameters) (string, error) { //url,parameters

	httpClient := http.Client{}

	var httpReq *http.Request

	dstUrl := srcUrl + "?" + parameters.Encode()

	httpReq, _ = http.NewRequest("GET", dstUrl, nil)

	println(httpReq)

	httpResp, err := httpClient.Do(httpReq)

	var str string

	if err == nil {
		defer httpResp.Body.Close()
		respBytes, _ := ioutil.ReadAll(httpResp.Body)
		str = string(respBytes)
		return str, nil
	}

	return "", err
}

func GetWithCookies(srcUrl string, parameters Parameters, cookies Parameters) (string, error) {

	var httpReq *http.Request

	dstUrl := srcUrl + "?" + parameters.Encode()

	httpReq, _ = http.NewRequest("GET", dstUrl, nil)
	httpReq.Header.Add("Cookie", cookies.CookieEncode())

	var httpResp *http.Response
	var err error
	httpClient := http.DefaultClient
	httpClient.Timeout = 最长允许响应时间
	var i int
	for {
		httpResp, err = httpClient.Do(httpReq)
		if err == nil {
			break
		} else if 重试次数 < 0 {
			log.Loglog("第", i+1, "次获取响应失败,无重试次数上限.", err.Error())
			i++
		} else if i == 重试次数 {
			log.Loglog("第", i+1, "次获取响应失败,到达重试次数上限.返回空响应.", err.Error())
			break
		} else {
			log.Loglog("第", i+1, "次获取响应失败,最多重试", 重试次数, "次.ERROR:", err.Error())
			i++
		}
	}

	var str string

	if err == nil {
		defer httpResp.Body.Close()
		respBytes, _ := ioutil.ReadAll(httpResp.Body)
		str = string(respBytes)
		return str, nil
	}

	return "", err

}

func Post(Url string, parameters Parameters) (string, error) {
	httpClient := http.Client{}

	var httpReq *http.Request

	httpReq, _ = http.NewRequest("POST", Url, bytes.NewReader([]byte(parameters.Encode())))
	httpReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	httpResp, err := httpClient.Do(httpReq)

	var str string

	if err == nil {
		defer httpResp.Body.Close()
		respBytes, _ := ioutil.ReadAll(httpResp.Body)
		str = string(respBytes)
		return str, nil
	}

	return "", err
}

func PostWithCookies(Url string, parameters Parameters) (string, error) {
	httpClient := http.Client{}

	var httpReq *http.Request

	httpReq, _ = http.NewRequest("POST", Url, bytes.NewReader([]byte(parameters.Encode())))
	httpReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Add("Cookie", parameters.CookieEncode())

	httpResp, err := httpClient.Do(httpReq)

	var str string

	if err == nil {
		defer httpResp.Body.Close()
		respBytes, _ := ioutil.ReadAll(httpResp.Body)
		str = string(respBytes)
		return str, nil
	}

	return "", err
}

/*
func GetResp(postUrl string, parameters Parameters) string {
	println(postUrl)
	println(parameters.Encode())
	httpClient := http.Client{}

	var httpReq *http.Request

	if parameters == nil {
		httpReq, _ = http.NewRequest("GET", postUrl, nil)
	} else {

		httpReq, _ = http.NewRequest("POST", postUrl, bytes.NewReader([]byte(parameters.Encode())))
		httpReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	var err error

	httpResp, err := httpClient.Do(httpReq)

	var str string

	if err == nil {
		defer httpResp.Body.Close()
		respBytes, _ := ioutil.ReadAll(httpResp.Body)
		str = string(respBytes)
		println(str)
		return str
	}
	str = "error"
	misc.Log(err.Error())

	return str

}

func GetRespByteSliceWithCookies(postUrl string, parameters Parameters, cookies Parameters) []byte {
	httpClient := http.Client{}

	var httpReq *http.Request

	if parameters == nil {
		httpReq, _ = http.NewRequest("GET", postUrl, nil)
	} else {

		httpReq, _ = http.NewRequest("POST", postUrl, bytes.NewReader([]byte(parameters.Encode())))
		httpReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	if cookies != nil {
		httpReq.Header.Add("Cookie", "")
		for _, cookie := range cookies {
			httpReq.AddCookie(&http.Cookie{
				Name:  cookie.Key,
				Value: cookie.Value,
			})
		}
	}

	var err error

	httpResp, err := httpClient.Do(httpReq)

	if err == nil {
		defer httpResp.Body.Close()
		respBytes, _ := ioutil.ReadAll(httpResp.Body)
		return respBytes
	}
	misc.Log(err.Error())
	return []byte("error")

}

func GetRespWithCookies(postUrl string, parameters Parameters, cookies Parameters) string {

	httpClient := http.Client{}

	var httpReq *http.Request

	if parameters == nil {
		httpReq, _ = http.NewRequest("GET", postUrl, nil)
	} else {

		httpReq, _ = http.NewRequest("POST", postUrl, bytes.NewReader([]byte(parameters.Encode())))
		httpReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	httpReq.Header.Add("Referer", "http://tieba.baidu.com/bawu2/platform/listMember?word=%D4%B5%D6%AE%BB")

	if cookies != nil {
		httpReq.Header.Add("Cookie", "")
		for _, cookie := range cookies {
			httpReq.AddCookie(&http.Cookie{
				Name:  cookie.Key,
				Value: cookie.Value,
			})
		}
	}

	var err error

	httpResp, err := httpClient.Do(httpReq)

	var str string

	if err == nil {
		defer httpResp.Body.Close()
		respBytes, _ := ioutil.ReadAll(httpResp.Body)
		str = string(respBytes)
		return str
	}
	str = "error"
	misc.Log(err.Error())
	return str

}

func GetRespByteSlice(postUrl string, parameters Parameters) []byte {

	httpClient := http.Client{}

	var httpReq *http.Request

	if parameters == nil {
		httpReq, _ = http.NewRequest("GET", postUrl, nil)
	} else {

		httpReq, _ = http.NewRequest("POST", postUrl, bytes.NewReader([]byte(parameters.Encode())))
		httpReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	var err error

	httpResp, err := httpClient.Do(httpReq)

	if err == nil {
		defer httpResp.Body.Close()
		respBytes, _ := ioutil.ReadAll(httpResp.Body)
		return respBytes
	}
	misc.Log(err.Error())
	return []byte("error")

}
*/
