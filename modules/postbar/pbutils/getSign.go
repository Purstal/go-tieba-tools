package pbutils

import (
	"encoding/json"
	"errors"
	//"fmt"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

type DataField struct {
	Content struct {
		Pid uint64 `json:"post_id"`
	} `json:"content"`
}

var PidNotMatchError error
var DataFieldNotFoundError error
var PostNotFoundError error
var SignSrcNotFoundError error

func init() {
	PidNotMatchError = errors.New("pid不匹配")
	PostNotFoundError = errors.New("未找到贴子")
	SignSrcNotFoundError = errors.New("找到签名档,但是未找到签名档的图片地址")
}

func GetUserSignUrl(tid, pid uint64) (string, error) {
	doc := TryGettingThreadDoc(tid, pid, 1)
	//println(len(doc.Text()))

	postSelection := doc.Find(`div.l_post`).First()

	dataFieldSrc, found := postSelection.Attr(`data-field`)
	if !found {
		return "", PostNotFoundError
	}

	var dataField DataField
	json.Unmarshal([]byte(dataFieldSrc), &dataField)

	if dataField.Content.Pid == pid {
		userSignSelection := postSelection.Find(`img.j_user_sign`)
		if userSignSelection.Length() == 1 {
			src, found := userSignSelection.Attr(`src`)
			if !found {
				return "", SignSrcNotFoundError
			}
			return src, nil
		} else {
			return "", nil
		}
	} else {
		return "", PidNotMatchError
	}

}

func TryGettingThreadDoc(tid, pid uint64, rn int) *goquery.Document {

	url := "http://tieba.baidu.com/p/" + strconv.FormatUint(tid, 10) + "?pid=" + strconv.FormatUint(pid, 10) + "&rn=" + strconv.Itoa(rn)
	for {
		doc, err := goquery.NewDocument(url)
		if err == nil && doc != nil {
			return doc
		}
	}
}
