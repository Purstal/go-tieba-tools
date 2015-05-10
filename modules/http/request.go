package http

import gohttp "net/http"

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/purstal/pbtools/modules/logs"
)

func useless() { fmt.Println() }

//http://www.crifan.com/go_language_http_do_post_pass_post_data/

var client *gohttp.Client
var RetryTimes int //重试次数

var ShutUp bool

func init() {
	client = &gohttp.Client{
		Timeout: time.Second * 10,
	}
	RetryTimes = 3
}

func Debug(str string) {
	//fmt.Println(str)
}

func Get(url string, parameters Parameters, cookies Cookies) ([]byte, error) {
	Debug(url)

	var httpReq *gohttp.Request

	if len(parameters) == 0 {
		httpReq, _ = gohttp.NewRequest("GET", url, nil)
	} else {
		httpReq, _ = gohttp.NewRequest("GET", url+"?"+parameters.Encode(), nil)
	}

	if len(cookies) != 0 {
		httpReq.Header.Add("Cookie", cookies.Encode())
	}

	for i := 0; ; {
		httpResp, err := client.Do(httpReq)
		defer func() {
			if err == nil && httpResp != nil && httpResp.Body != nil {
				httpResp.Body.Close()
			}

		}()

		if err == nil {
			bs, _ := ioutil.ReadAll(httpResp.Body)
			return bs, nil
		} else if RetryTimes < 0 {
			if !ShutUp {
				logs.Fatal("第", i+1, "次获取响应失败,无重试次数上限.", err.Error())
			}
			i++
		} else if i == RetryTimes {
			if !ShutUp {
				logs.Fatal("第", i+1, "次获取响应失败,到达重试次数上限.", err.Error())
			}
			return []byte(""), err
		} else {
			if !ShutUp {
				logs.Fatal("第", i+1, "次获取响应失败,最多重试", RetryTimes, "次.ERROR:", err.Error())
			}
			i++
		}
	}

}

func Post(url string, parameters Parameters) ([]byte, error) {
	Debug(url)

	var httpReq *gohttp.Request

	if len(parameters) == 0 {
		httpReq, _ = gohttp.NewRequest("POST", url, nil)
	} else {
		httpReq, _ = gohttp.NewRequest("POST", url, bytes.NewReader([]byte(parameters.Encode())))
	}

	httpReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	for i := 0; ; {
		httpResp, err := client.Do(httpReq)
		defer func() {
			if err == nil && httpResp != nil && httpResp.Body != nil {
				httpResp.Body.Close()
			}

		}()

		if err == nil {
			bs, _ := ioutil.ReadAll(httpResp.Body)
			return bs, nil
		} else if RetryTimes < 0 {
			if !ShutUp {
				logs.Fatal("第", i+1, "次获取响应失败,无重试次数上限.", err.Error())
			}
			i++
		} else if i == RetryTimes {
			if !ShutUp {
				logs.Fatal("第", i+1, "次获取响应失败,到达重试次数上限.", err.Error())
			}
			return []byte(""), err
		} else {
			if !ShutUp {
				logs.Fatal("第", i+1, "次获取响应失败,最多重试", RetryTimes, "次.ERROR:", err.Error())
			}
			i++
		}
	}
}

func PostAdv(url string, parameters Parameters, cookies Cookies, f func(*gohttp.Request)) ([]byte, error) {
	Debug(url)

	var httpReq *gohttp.Request

	if len(parameters) == 0 {
		httpReq, _ = gohttp.NewRequest("POST", url, nil)
	} else {
		httpReq, _ = gohttp.NewRequest("POST", url, bytes.NewReader([]byte(parameters.Encode())))
	}

	if len(cookies) != 0 {
		httpReq.Header.Add("Cookie", cookies.Encode())
	}

	httpReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	f(httpReq)

	for i := 0; ; {
		httpResp, err := client.Do(httpReq)
		defer func() {
			if err == nil && httpResp != nil && httpResp.Body != nil {
				httpResp.Body.Close()
			}

		}()

		if err == nil {
			bs, _ := ioutil.ReadAll(httpResp.Body)
			return bs, nil
		} else if RetryTimes < 0 {
			if !ShutUp {
				logs.Fatal("第", i+1, "次获取响应失败,无重试次数上限.", err.Error())
			}
			i++
		} else if i == RetryTimes {
			if !ShutUp {
				logs.Fatal("第", i+1, "次获取响应失败,到达重试次数上限.", err.Error())
			}
			return []byte(""), err
		} else {
			if !ShutUp {
				logs.Fatal("第", i+1, "次获取响应失败,最多重试", RetryTimes, "次.ERROR:", err.Error())
			}
			i++
		}
	}
}

func GetAdv(url string, parameters Parameters, cookies Cookies, f func(*gohttp.Request)) ([]byte, error) {
	Debug(url)

	var httpReq *gohttp.Request

	if len(parameters) == 0 {
		httpReq, _ = gohttp.NewRequest("GET", url, nil)
	} else {
		httpReq, _ = gohttp.NewRequest("GET", url+"?"+parameters.Encode(), nil)
	}

	if len(cookies) != 0 {
		httpReq.Header.Add("Cookie", cookies.Encode())
	}

	f(httpReq)

	for i := 0; ; {
		httpResp, err := client.Do(httpReq)
		//fmt.Println(err)
		//fmt.Println(httpResp, err)
		defer func() {
			if err == nil && httpResp != nil && httpResp.Body != nil {
				httpResp.Body.Close()
			}

		}()
		if err == nil {
			bs, _ := ioutil.ReadAll(httpResp.Body)
			return bs, nil
		} else if RetryTimes < 0 {
			if !ShutUp {
				logs.Fatal("第", i+1, "次获取响应失败,无重试次数上限.", err.Error())
			}
			i++
		} else if i == RetryTimes {
			if !ShutUp {
				logs.Fatal("第", i+1, "次获取响应失败,到达重试次数上限.", err.Error())
			}
			return []byte(""), err
		} else {
			if !ShutUp {
				logs.Fatal("第", i+1, "次获取响应失败,最多重试", RetryTimes, "次.ERROR:", err.Error())
			}
			i++
		}
	}

}
