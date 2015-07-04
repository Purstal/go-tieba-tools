package utils

type ThreadManager struct {
	RequireChan, DoChan, FinishChan chan bool

	MaxThreadNumber,
	TotalTaskCount int

	ActiveCount,
	WaitingCount,
	FinishedCount int

	AllTaskFinishedChan chan bool
}

func NewThreadManager(maxThreadNumber, totalTaskCount int) *ThreadManager {
	var m ThreadManager
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
