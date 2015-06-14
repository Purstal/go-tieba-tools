package post_deleter

import (
	"github.com/purstal/pbtools/modules/postbar"
	"github.com/purstal/pbtools/modules/postbar/apis"
)

func (d *PostDeleter) DeletePost(from string, account *postbar.Account, tid, pid, spid, uid uint64, reason string) {

	if account.BDUSS == "" {
		d.Logger.Warn("BDUSS为空,忽略删贴请求.")
		return
	}

	var op_pid uint64
	if spid != 0 {
		op_pid = spid
	} else {
		op_pid = pid
	}

	prefix := MakePrefix(nil, tid, pid, spid, uid)
	d.OpLogger.Info(prefix, from, "删贴:", reason, ".")

	for i := 0; ; i++ {
		err, pberr := apis.DeletePost(account, op_pid)
		if err == nil && (pberr == nil || pberr.ErrorCode == 0) {
			return
		} else if i < 3 {
			d.OpLogger.Error(prefix, "删贴失败,将最多尝试三次:", err, pberr, ".")
		} else {
			d.OpLogger.Error(prefix, "删贴失败,放弃:", err, pberr, ".")
			return
		}
	}
}

func (d *PostDeleter) DeleteThread(from string, account *postbar.Account, tid, pid, uid uint64, reason string) {

	if account.BDUSS == "" {
		d.Logger.Warn("BDUSS为空,忽略删主题请求.")
		return
	}

	prefix := MakePrefix(nil, tid, pid, 0, uid)
	d.OpLogger.Info(prefix, from, "删主题:", reason, ".")

	for i := 0; ; i++ {
		err, pberr := apis.DeleteThread(account, tid)
		if err == nil && (pberr == nil || pberr.ErrorCode == 0) {
			return
		} else if i < 3 {
			d.OpLogger.Error(prefix, "删主题失败,将最多尝试三次:", err, pberr, ".")
		} else {
			d.OpLogger.Error(prefix, "删主题失败,放弃:", err, pberr, ".")
			return
		}
	}
}

func (d *PostDeleter) BanID(from string, BDUSS string, userName string, fid, tid, pid, uid uint64, day int, loggedReason, givedReason string) {

	if BDUSS == "" {
		d.Logger.Warn("BDUSS为空,忽略封禁请求.")
		return
	}

	prefix := MakePrefix(nil, tid, pid, 0, uid)
	d.OpLogger.Info(prefix, from, "封禁:", loggedReason+"(给出: "+givedReason+")", ".")

	for i := 0; ; i++ {
		err, pberr := apis.BlockIDWeb(BDUSS, fid, userName, pid, day, givedReason)
		if err == nil && (pberr == nil || pberr.ErrorCode == 0) {
			return
		} else if i < 3 {
			d.OpLogger.Error(prefix, "封禁失败,将最多尝试三次:", err, pberr, ".")
		} else {
			d.OpLogger.Error(prefix, "封禁失败,放弃:", err, pberr, ".")
			return
		}
	}

}
