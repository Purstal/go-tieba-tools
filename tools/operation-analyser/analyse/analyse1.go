package analyse

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	analyser "github.com/purstal/pbtools/tool-cores/operation-analyser"

	"github.com/purstal/pbtools/tools/operation-analyser/csv"
)

func Analyse1(datas []analyser.DayData) {
	var bawuTotal = make(map[string]int)
	var hourCounts = make([][24]map[string]int, len(datas))

	for i, data := range datas {
		for j := 0; j < 24; j++ {
			hourCounts[i][j] = make(map[string]int)
		}
		for _, log := range data.Logs {
			hourCounts[i][log.OperateTime.Hour][log.Operator]++
			bawuTotal[log.Operator]++
		}
	}
	var records sorter

	for bawu, count := range bawuTotal {
		records = append(records, record{bawu, count})
	}

	sort.Sort(records)

	var table = make([][]string, len(datas)*30+1)

	for i, _ := range table {
		table[i] = make([]string, len(records)+1)
	}

	for i, record := range records {
		table[0][i+1] = record.userName
	}

	for day, counts := range hourCounts {
		t := time.Unix(datas[day].Time, 0)
		for hour, countMap := range counts {
			row := &table[1+24*day+hour]
			(*row)[0] = t.Format("2006-01-02 15:00")
			t = t.Add(time.Hour)
			for i, record := range records {
				(*row)[1+i] = strconv.Itoa(countMap[record.userName])
			}
		}
	}

	f, _ := os.Create("result.csv")
	w := csv.NewWriter(f)
	w.WriteAll(table)
	w.Flush()
	fmt.Println(records)

}

type record struct {
	userName string
	count    int
}

type sorter []record

func (s sorter) Less(i, j int) bool {
	return s[i].count > s[j].count
}

func (s sorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sorter) Len() int {
	return len(s)
}
