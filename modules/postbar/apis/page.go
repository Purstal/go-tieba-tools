package apis

import (
	"strconv"

	"github.com/purstal/pbtools/modules/http"
	"github.com/purstal/pbtools/modules/postbar"
)

func RGetForum(acc *postbar.Account, kw string, rn,
	pn int) ([]byte, error) {
	var parameters http.Parameters

	parameters.Add("kw", kw)
	parameters.Add("rn", strconv.Itoa(rn))
	parameters.Add("pn", strconv.Itoa(pn))
	//is_good

	postbar.ProcessParams(&parameters, acc)

	return http.Post(`http://c.tieba.baidu.com/c/f/frs/page`, parameters)

}

func RGetThread(acc *postbar.Account, tid uint64, mark bool, pid uint64, pn, rn int,
	withFloor, seeLz, r bool) ([]byte, error) {
	var parameters http.Parameters

	parameters.Add("kz", strconv.FormatUint(tid, 10))

	if mark {
		parameters.Add("mark", "1") //1:定位楼层
	}

	if pid != 0 {
		parameters.Add("pid", strconv.FormatUint(pid, 10)) //用于定位楼中楼
	}

	if rn != 30 {
		parameters.Add("rn", strconv.Itoa(rn)) //至少为2
	}
	if pn != 0 {
		parameters.Add("pn", strconv.Itoa(pn))
	}
	//parameters.Add("back", "0")       //1:不会显示一楼
	if r {
		parameters.Add("r", "1") //1:倒序查看
	}
	if seeLz {
		parameters.Add("lz", "1") //1:只看楼主
	}
	if withFloor {
		parameters.Add("with_floor", "1") //0:不带楼中楼;省缺为0
	}
	//parameters.Add("last", "1")       //?

	postbar.ProcessParams(&parameters, acc)

	return http.Post(`http://c.tieba.baidu.com/c/f/pb/page`, parameters)

}

func RGetFloor(acc *postbar.Account, tid uint64,
	isComment bool, id uint64, pn int) ([]byte, error) {
	var parameters http.Parameters
	parameters.Add("from", "tieba")
	parameters.Add("kz", strconv.FormatUint(tid, 10))
	if isComment {
		parameters.Add("spid", strconv.FormatUint(id, 10))
	} else {
		parameters.Add("pid", strconv.FormatUint(id, 10))
	}
	if pn != 0 {
		parameters.Add("pn", strconv.Itoa(pn))
	}

	postbar.ProcessParams(&parameters, acc)

	resp, err := http.Post(`http://c.tieba.baidu.com/c/f/pb/floor`, parameters)
	return resp, err
}
