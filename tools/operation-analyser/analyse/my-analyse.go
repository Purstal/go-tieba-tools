package analyse

import (
	"fmt"
	"os"
	"sort"
	"time"

	analyser "github.com/purstal/pbtools/tool-cores/operation-analyser"
)

func Analyse2(datas []analyser.DayData) {
	var sentenceMap = make(map[string]int)
	var deletedByIamunknown = make(map[int]bool)
	for _, data := range datas {
		for _, log := range data.Logs {
			if log.Operator == "iamunknown" {
				deletedByIamunknown[log.PID] = true
			}
		}
	}
	for _, data := range datas {
		for _, log := range data.Logs {
			if log.Operator == "iamunknown" || log.OperateType != analyser.OpType_Delete {
				continue
			}
			var sentence = log.Text
			/*for _, sentence := range split([]rune(log.Text), []rune(` 	.,!?:;'"。，！？：；‘’“”`)) {*/
			if !deletedByIamunknown[log.PID] {
				sentenceMap[string(sentence)]++
			}
			/*}*/
		}
	}

	var sentenceSlice []record2

	for k, v := range sentenceMap {
		sentenceSlice = append(sentenceSlice, record2{k, v})
	}

	sort.Sort(sorter2(sentenceSlice))

	var f, _ = os.Create("" + time.Now().Format("result-20060102-150405.txt"))

	for _, sentence := range sentenceSlice {
		if sentence.time < 2 {
			break
		}
		fmt.Fprintln(f, sentence.time, ":", sentence.text)
	}
	f.Close()
}

type record2 struct {
	text string
	time int
}

func split(s, seps []rune) [][]rune {
	var result [][]rune
	var last int
	for i, r := range s {
		for _, sep := range seps {
			if r == sep {
				if last != i {
					result = append(result, s[last:i])
				}
				last = i + 1
				continue
			}
		}
	}
	return result
}

type sorter2 []record2

func (s sorter2) Less(i, j int) bool {
	return s[i].time > s[j].time
}

func (s sorter2) Len() int {
	return len(s)
}

func (s sorter2) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
