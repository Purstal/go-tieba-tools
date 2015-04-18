package post_finder

//type debugTaskLog interface{}

//type debugTaskLogs []debugTaskLog
type debugTaskLogs []string

func (logs *debugTaskLogs) Log(x string) {
	*logs = append(*logs, x)
}
