package collects

type ThreadRecord struct {
	Title    string
	Tid      uint64
	PostTime int64
	Author   string
	Abstract []interface{}
}
