package save_to_html

import (
	"fmt"
	//"html/template"
	"io"
	"time"
)

/*
var T *template.Template

func init() {
	T = template.New("Why I Need A Name?")
	var err error
	T, err = T.Parse(RECORD_TEMPLATE)
	if err != nil {
		panic(err)
	}
}
*/

func Write(w io.Writer, records []Record, timeStr string) {
	var now = time.Now()
	fmt.Fprintf(w, MAIN_TEMPLATE_FIRST_HALF, fmt.Sprintf("%d", now.UnixNano()), timeStr)
	for _, r := range records {
		fmt.Fprintf(w, RECORD_TEMPLATE, r.Tid, r.Title,
			r.Tid, r.Tid, r.Tid, r.Author, r.Tid, r.Tid,
			r.Time, r.Abstract)
	}
	w.Write(MAIN_TEMPLATE_SECOND_HALF_bytes)
}

func ExtractAbstract(contents []interface{}) string {
	if len(contents) == 0 {
		return ""
	}

	var str = ""

	for _, _content := range contents {
		content, ok1 := _content.(map[string]interface{})
		t, ok2 := content["type"].(string)
		if !ok1 {
			t = "解析失败"
		} else if !ok2 {
			t = "缺少类型"
		}

		switch t {
		case "0": //文字
			if content["text"] != nil {
				str += fmt.Sprintf(`<li>文字: %s</li>`, fmt.Sprint(content["text"]))
			} else {
				str += fmt.Sprintf(`<li>文字: 解析失败(%s)</li>`, fmt.Sprint(content))
			}
		case "3": //图片
			if content["big_pic"] != nil {
				str += fmt.Sprintf(`<li>图片: <img src="%s" width=240 height=160></img></li>`, fmt.Sprint(content["big_pic"]))
			} else {
				str += fmt.Sprintf(`<li>图片: 解析失败(%s)</li>`, fmt.Sprint(content))
			}
		case "5": //视频
			if content["vhsrc"] != nil {
				str += fmt.Sprintf(`<li>视频: %s`, fmt.Sprint(content["vhsrc"]))
				if content["vsrc"] != nil {
					str += fmt.Sprint(`(%s)`, content["vsrc"])
				}
				str += `</li>`
			} else if content["vsrc"] != nil {
				str += fmt.Sprintf(`<li>视频: %s</li>`, fmt.Sprint(content["vsrc"]))
			} else {
				str += fmt.Sprintf(`<li>视频: 解析失败(%s)</li>`, fmt.Sprint(content))
			}
		case "6":
			if content["src"] != nil {
				str += fmt.Sprintf(`<li>音乐: %s</li>`, fmt.Sprint(content["src"]))
			} else {
				str += fmt.Sprintf(`<li>音乐: 解析失败(%s)</li>`, fmt.Sprint(content))
			}
		case "解析失败":
			str += fmt.Sprintf(`<li>未知: 解析失败(%s)</li>`, fmt.Sprint(content))
		case "缺少类型":
			str += fmt.Sprintf(`<li>未知: 缺少类型(%s)</li>`, fmt.Sprint(content))
		default:
			str += fmt.Sprintf(`<li>未知: 未知类型(%s)(%s)</li>`, t, fmt.Sprint(content))
		}
	}

	return str
}
