package collects

import (
	"sort"
)

func (theMap ThreadMap) ToSortedSlice() []Thread {
	var slice = make([]Thread, len(theMap))
	var i = 0
	for _, v := range theMap {
		slice[i] = v
		i++
	}
	sort.Sort(ThreadSliceSorter(slice))
	return slice
}

type ThreadSliceSorter []Thread

func (s ThreadSliceSorter) Less(i, j int) bool {
	return s[i].Tid < s[j].Tid
}

func (s ThreadSliceSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ThreadSliceSorter) Len() int {
	return len(s)
}
