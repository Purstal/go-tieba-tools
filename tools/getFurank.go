package main

import (
	"io"
	"net/http"
	"os"
	"time"

	"github.com/purstal/pbtools/modules/misc"
)

func main() {
	if len(os.Args) == 1 {
		return
	}
	var forumName = misc.ToGBK(os.Args[1])
	var resp *http.Response
	for {
		var err error
		resp, err = http.Get("http://tieba.baidu.com/f/like/furank?kw=" + forumName + "&pn=1")
		if err == nil {
			break
		}
	}
	f, _ := os.Create(time.Now().Format(forumName + "-20060102-150405.html"))
	io.Copy(f, resp.Body)
}
