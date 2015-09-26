package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	pbhttp "github.com/purstal/go-tieba-base/simple-http"
	"github.com/purstal/go-tieba-base/tieba"
	"github.com/purstal/go-tieba-base/tieba/apis"
)

type MyMux struct{}

func useless() {
	fmt.Println()
}

func main() {

	var mux = &MyMux{}

	var port int

	if len(os.Args) != 1 {
		var err error
		port, err = strconv.Atoi(os.Args[1])
		if err != nil {
			fmt.Println("输入的端口有误:", err)
			port = 33120
		}
	} else {
		port = 33120
	}

	fmt.Println("使用端口:", port)

	if err := http.ListenAndServe(":"+strconv.Itoa(port), mux); err != nil {
		fmt.Println(err)
		return
	}

	<-make(chan bool)
}

func (mux *MyMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "/favicon.ico" {
		return
	}

	r.ParseForm()

	var form = make(map[string]string)

	for key, value := range r.Form {
		if len(value) != 0 {
			form[key] = value[0]
		}
	}

	var fmt_json, debug_output bool

	if _, exist := form["fmt-json"]; exist {
		delete(form, "fmt-json")
		fmt_json = true
	}

	if _, exist := form["console-debug-output"]; exist {
		delete(form, "console-debug-output")
		debug_output = true
	}

	var account *postbar.Account

	switch form[("client")] {
	case "", "Win8":
		account = postbar.NewDefaultWindows8Account("")
	case "Andr":
		account = postbar.NewDefaultAndroidAccount("")
	case "nil":
		account = &postbar.Account{}
	default:
		account = &postbar.Account{}
	}

	if net_type, exist := form["net_type"]; exist {
		delete(form, "net_type")
		account.NetType = net_type
	}
	if _client_type, exist := form["_client_type"]; exist {
		delete(form, "_client_type")
		account.ClientType = _client_type
	}
	if _client_id, exist := form["_client_id"]; exist {
		delete(form, "_client_id")
		account.ClientID = _client_id
	}
	if _client_version, exist := form["_client_version"]; exist {
		delete(form, "_client_version")
		account.ClientVersion = _client_version
	}
	if _phone_imei, exist := form["_phone_imei"]; exist {
		delete(form, "_phone_imei")
		account.PhoneIMEI = _phone_imei
	}

	delete(form, "client")

	if BDUSS, exist := form[("BDUSS")]; exist {
		delete(form, "BDUSS")
		account.BDUSS = BDUSS
	}

	var parameters pbhttp.Parameters

	if _, exist := form["require-tbs"]; exist {
		delete(form, "require-tbs")
		for {
			tbs, err := apis.GetTbsWeb(account.BDUSS)
			if err == nil {
				parameters.Add("tbs", tbs)
				break
			}
		}
	}

	if _, exist := form["require-cuid"]; exist {
		delete(form, "require-cuid")
		cuid := postbar.GenCUID("", account.PhoneIMEI)
		parameters.Add("cuid", cuid)
	}

	for key, value := range form {
		parameters.Add(key, value)
	}

	postbar.ProcessParams(&parameters, account)

	if debug_output {
		fmt.Println(r.URL.Path, parameters.Encode(), "\n")

	}

	resp, err := pbhttp.Post("http://c.tieba.baidu.com"+r.URL.Path, parameters)
	var ERROR struct {
		UN_ERROR string
	}
	if err != nil {
		ERROR.UN_ERROR = err.Error()
		data, _ := json.Marshal(ERROR)
		w.Write(data)
		return
	}
	if fmt_json {
		var data interface{}
		json.Unmarshal(resp, &data)
		resp, err = json.MarshalIndent(data, "", "  ")
		if err != nil {
			ERROR.UN_ERROR = err.Error()
			data, _ := json.Marshal(ERROR)
			w.Write(data)
			return
		}
	}
	w.Write(resp)

}
