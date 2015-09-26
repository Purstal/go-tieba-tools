package main

import (
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/purstal/go-tieba-base/misc"
)

func main() {
	if len(os.Args) == 1 {
		panic("Incorrent Usage.")
		return
	}
	var to = 1
	if len(os.Args) == 3 {
		var err error
		to, err = strconv.Atoi(os.Args[2])
		if err != nil {
			panic(err)
		} else if to < 1 {
			panic("Incorrent Page Number.")
		}
	}
	var forumName = misc.ToGBK(os.Args[1])
	var dir = "furank/" + forumName + "/"
	os.MkdirAll(dir, 0644)
	var now = time.Now()
	for pn := 1; pn <= to; pn++ {
		var resp *http.Response
		for {
			var err error
			resp, err = http.Get("http://tieba.baidu.com/f/like/furank?kw=" + forumName + "&pn=" + strconv.Itoa(pn))
			if err == nil {
				break
			}
		}
		f, err := os.Create(dir + now.Format("20060102-150405-") + strconv.Itoa(pn) + ".html")
		if err != nil {
			panic(err)
		}
		io.Copy(f, resp.Body)
	}

}
