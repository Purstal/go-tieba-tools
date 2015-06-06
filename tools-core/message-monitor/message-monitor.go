package message_monitor

import (
	//"fmt"
	"time"

	"github.com/purstal/pbtools/modules/logs"
	"github.com/purstal/pbtools/modules/pberrors"
	"github.com/purstal/pbtools/modules/postbar"

	"github.com/purstal/pbtools/tools-core/utils/action"

	//"github.com/purstal/pbtools/modules/postbar/floor-andr-6.1.3"
	//"github.com/purstal/pbtools/modules/postbar/thread-win8-1.5.0.0"

	"github.com/purstal/pbtools/modules/postbar/message"
)

type MessageMonitor struct {
	Interval    time.Duration
	MessageChan chan message.ReplyMessage

	actChan chan action.Action
}

const (
	Stop = iota
)

func NewMessageMonitor(logger *logs.Logger, interval time.Duration, acc *postbar.Account, lastFoundPid uint64) *MessageMonitor {
	var monitor MessageMonitor
	monitor.Interval = interval
	monitor.MessageChan = make(chan message.ReplyMessage)
	monitor.actChan = make(chan action.Action)

	go func() {

		ticker := time.NewTicker(monitor.Interval)
		for {

			msgs, _lastFoundPid, err, pberr := checkReply(acc, lastFoundPid)
			if len(msgs) > 0 {
				if err != nil {
					logger.Error("无法获取消息提醒", err)
					continue
				} else if pberr != nil && pberr.ErrorCode != 0 {
					logger.Error("无法获取消息提醒", pberr)
					continue
				}
				lastFoundPid = _lastFoundPid
				for _, msg := range msgs {
					monitor.MessageChan <- msg
				}
			}

			select {
			case <-ticker.C:
			case act := <-monitor.actChan:
				switch act.Pattern {
				case Stop:
					ticker.Stop()
					//logs.Debug("喵")
					close(monitor.MessageChan)
					//logs.Debug("喵喵")
					close(monitor.actChan)
					//logs.Debug("将解引赋值:", lastFoundPid)
					*(act.Param.(*chan uint64)) <- lastFoundPid
					return
				}
			}
		}
	}()
	return &monitor
}

func (m *MessageMonitor) Stop() uint64 {
	var lastFoundPidChan = make(chan uint64)
	m.actChan <- action.Action{Stop, &lastFoundPidChan}
	lastFoundPid := <-lastFoundPidChan
	//logs.Debug("解引所赋值:", lastFoundPid)
	return lastFoundPid
}

func checkReply(acc *postbar.Account, lastFoundPid uint64) ([]message.ReplyMessage, uint64, error, *pberrors.PbError) {

	var msgs []message.ReplyMessage
	var err error
	var pberr *pberrors.PbError

	for i := 0; i < 10; i++ {
		msgs, err, pberr = message.GettReplyMessageStruct(acc)
		if err == nil {
			break
		}
	}

	if err != nil || len(msgs) < 0 || (pberr != nil && pberr.ErrorCode != 0) {
		return nil, lastFoundPid, err, pberr
	}

	for i, msg := range msgs {
		if msg.Pid <= lastFoundPid {
			return msgs[:i], msgs[0].Pid, nil, nil
		}
	}

	return msgs, msgs[0].Pid, nil, nil

}
