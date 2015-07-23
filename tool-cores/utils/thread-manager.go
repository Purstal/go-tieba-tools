package utils

type LimitTaskManager struct {
	RequireChan chan bool
	DoChan      chan bool
	FinishChan  chan bool

	MaxThreadNumber,
	TotalTaskCount int

	ActiveCount,
	WaitingCount,
	FinishedCount int

	AllTaskFinishedChan chan bool
}

func NewLimitTaskManager(maxThreadNumber, totalTaskCount int) *LimitTaskManager {
	var m LimitTaskManager
	m.RequireChan, m.DoChan, m.FinishChan = make(chan bool), make(chan bool), make(chan bool)
	m.AllTaskFinishedChan = make(chan bool)
	m.MaxThreadNumber = maxThreadNumber
	m.TotalTaskCount = totalTaskCount
	go func() {
		for {
			select {
			case <-m.RequireChan:
				if m.ActiveCount < m.MaxThreadNumber {
					m.DoChan <- true
					m.ActiveCount++
				} else {
					m.WaitingCount++
				}
			case <-m.FinishChan:
				m.FinishedCount++
				if m.FinishedCount == m.TotalTaskCount {
					for i := 0; i < m.WaitingCount; i++ {
						m.DoChan <- false
					}
					close(m.RequireChan)
					close(m.DoChan)
					close(m.FinishChan)
					m.AllTaskFinishedChan <- true
					close(m.AllTaskFinishedChan)
					return
				}
				if m.WaitingCount > 0 {
					m.DoChan <- true
					m.WaitingCount--
				} else {
					m.ActiveCount--
				}
			}
		}
	}()
	return &m
}

type UnlimitTaskManager struct {
	DemandChan         chan struct{} //要求进行任务
	DoChan             chan bool     //若为false, 所有任务完成
	TaskFinishesChan   chan struct{} //汇报任务完成
	AllTasksFinishChan chan struct{} //汇报所有任务完成

	IsAllTasksFinished bool

	MaxThreadNumber int

	ActiveCount,
	WaitingCount int

	WaitForFinishChan chan struct{} //阻塞直至所有任务完成
}

func NewUnlimitTaskManager(maxThreadNumber int) *UnlimitTaskManager {
	var m UnlimitTaskManager
	m.DemandChan, m.DoChan, m.TaskFinishesChan, m.AllTasksFinishChan = make(chan struct{}), make(chan bool), make(chan struct{}), make(chan struct{})
	m.WaitForFinishChan = make(chan struct{})
	m.MaxThreadNumber = maxThreadNumber
	go func() {
		for {
			select {
			case <-m.DemandChan:
				if m.ActiveCount < m.MaxThreadNumber {
					m.DoChan <- true
					m.ActiveCount++
				} else {
					m.WaitingCount++
				}
			case <-m.TaskFinishesChan:
				if m.IsAllTasksFinished {
					for i := 0; i < m.WaitingCount; i++ {
						m.DoChan <- false
					}
					for i := 0; i < m.ActiveCount-1; i++ {
						<-m.TaskFinishesChan
					}
					close(m.DemandChan)
					close(m.DoChan)
					close(m.TaskFinishesChan)
					close(m.AllTasksFinishChan)
					m.WaitForFinishChan <- struct{}{}
					close(m.WaitForFinishChan)
					return
				}
				if m.WaitingCount > 0 {
					m.DoChan <- true
					m.WaitingCount--
				} else {
					m.ActiveCount--
				}
			case <-m.AllTasksFinishChan:
				m.IsAllTasksFinished = true
			}
		}
	}()
	return &m
}
