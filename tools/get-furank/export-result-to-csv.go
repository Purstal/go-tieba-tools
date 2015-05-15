package main

import (
	//"encoding/json"
	//"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/purstal/pbtools/modules/misc"
	"github.com/purstal/pbtools/tools/operation-analyser/csv"
)

func main() {
	const dirName = "minecraft/"

	dirFis, err1 := ioutil.ReadDir(dirName)
	if err1 != nil {
		panic(err1)
	}

	var time_userName_exp_map = make(map[int64]map[string]int)
	type maxAndMin struct {
		max, min int
	}
	var userMap = make(map[string]maxAndMin)
	var timeMap = make(map[int64]bool)

	var rx = regexp.MustCompile(`(\d*-\d*)`)

	for _, fi := range dirFis {
		var timeStr = rx.FindString(fi.Name())
		var t, err1 = time.Parse("20060102-150405", timeStr)
		if err1 != nil {
			panic(err1)
		}
		timeMap[t.Unix()] = true

		var records map[string]int

		if _r, found := time_userName_exp_map[t.Unix()]; found {
			records = _r
		} else {
			records = make(map[string]int)
			time_userName_exp_map[t.Unix()] = records
		}

		var f, err2 = os.Open(dirName + fi.Name())
		if err2 != nil {
			panic(err2)
		}
		defer f.Close()

		var doc, err3 = goquery.NewDocumentFromReader(f)
		if err3 != nil {
			panic(err3)
		}

		doc.Find(`tr.drl_list_item`).Each(func(i int, item *goquery.Selection) {
			//fmt.Println(misc.FromGBK(item.Find(`td.drl_item_name`).Text()), item.Find(`td.drl_item_exp`).Text())
			var userName = misc.FromGBK(item.Find(`td.drl_item_name`).Text())
			var exp, _ = strconv.Atoi(item.Find(`td.drl_item_exp`).Text())
			if m, found := userMap[userName]; found {
				if m.max < exp {
					userMap[userName] = maxAndMin{exp, m.min}
				} else if m.min > exp {
					userMap[userName] = maxAndMin{m.max, exp}
				}
			} else {
				userMap[userName] = maxAndMin{exp, exp}
			}

			records[userName] = exp
		})
	}

	var userSlice []record

	for userName, m := range userMap {
		userSlice = append(userSlice, record{userName, m.max - m.min})
	}

	var timeSlice []int64

	for t, _ := range timeMap {
		timeSlice = append(timeSlice, t)
	}

	sort.Sort(Sorter(userSlice))
	sort.Sort(StupidInt64Sorter(timeSlice))

	var table = make([][]string, len(timeSlice)+1)
	table[0] = make([]string, len(userSlice)+1)
	for i := 1; i < len(userSlice)+1; i++ {
		table[0][i] = userSlice[i-1].userName
	}

	for i := 1; i < len(table); i++ {
		table[i] = make([]string, len(userSlice)+1)
		table[i][0] = time.Unix(timeSlice[i-1], 0).Format("2006-01-02 15:") + "00"
		for j := 1; j < len(userSlice)+1; j++ {
			var exp = time_userName_exp_map[timeSlice[i-1]][userSlice[j-1].userName]
			if exp != 0 {
				table[i][j] = strconv.Itoa(exp)
			} else {
				table[i][j] = "-"
			}

		}
	}

	var out, err2 = os.Create("result.csv")
	if err2 != nil {
		panic(err2)
	}
	defer out.Close()

	var w = csv.NewWriter(out)
	w.WriteAll(table)
	w.Flush()

}

type record struct {
	userName string
	diff     int
}

type Sorter []record

func (s Sorter) Less(i, j int) bool {
	return s[i].diff > s[j].diff
}

func (s Sorter) Len() int {
	return len(s)
}

func (s Sorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type StupidInt64Sorter []int64

func (s StupidInt64Sorter) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s StupidInt64Sorter) Len() int {
	return len(s)
}

func (s StupidInt64Sorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
