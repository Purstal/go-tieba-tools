package caozuoliang

import "github.com/purstal/pbtools/modules/http"
import "github.com/purstal/pbtools/tool-core/operation-analyser/old/log"

import (
	"strconv"
	"time"
)

func ListPostLog_op(BDUSS, word, svalue string, op_type OPtype, begin, end *time.Time, pn int) (string, error) {

	bt := time.Now()
	var parameters http.Parameters

	parameters.Add("word", word)
	parameters.Add("op_type", string(op_type))
	parameters.Add("stype", "op_uname")                          //post_uname:发贴人;op_uname:操作人
	parameters.Add("svalue", svalue)                             //utf8-url-encoding
	parameters.Add("date_type", "on")                            //限定时间范围
	parameters.Add("begin", strconv.FormatInt(begin.Unix(), 10)) //起始时间戳
	parameters.Add("end", strconv.FormatInt(end.Unix(), 10))     //结束时间戳
	parameters.Add("pn", strconv.Itoa(pn))

	var cookies http.Cookies

	cookies.Add("BDUSS", BDUSS)

	resp, err := http.Get("http://tieba.baidu.com/bawu2/platform/listPostLog", parameters, cookies)

	log.Log233("响应时间", time.Now().Sub(bt).String())
	return string(resp), err
}
