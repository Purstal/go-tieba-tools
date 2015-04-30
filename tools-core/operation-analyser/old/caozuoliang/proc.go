package caozuoliang

func ProcessData(ch <-chan *PostLog) []*PostLog {
	var pls []*PostLog
	for pl := range ch {
		//fmt.Println("ProcessData", pl)
		pls = append(pls, pl)

	}
	//log2("ProcessData:break")

	return pls
}
