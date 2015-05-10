package thread

import (
	"math"
	"strconv"
	"time"

	"github.com/purstal/pbtools/modules/http"
	"github.com/purstal/pbtools/modules/pberrors"
	"github.com/purstal/pbtools/modules/postbar"
)

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

func GetThreadStruct(acc *postbar.Account, tid uint64, mark bool, pid uint64, pn, rn int,
	withFloor, seeLz, r bool) (*ThreadPage, []ThreadPagePost,
	*ThreadPageExtra, error, *pberrors.PbError) {

	otp, err, pberr := GetOriginalThreadStruct(acc, tid, mark, pid, pn, rn, withFloor, seeLz, r)

	if err != nil {
		return nil, nil, nil, err, nil
	}
	var tpe ThreadPageExtra
	tpe.CurrentPage = otp.Page.CurrentPage
	tpe.TotalPage = otp.Page.TotalPage
	tpe.ServerTime = time.Unix(otp.Time, 0)
	if pberr != nil {
		return nil, nil, &tpe, nil, pberr
	}

	var tp ThreadPage

	var postList []ThreadPagePost = make([]ThreadPagePost, len(otp.PostList))

	tp.Tid = otp.Thread.ID
	tp.Title = otp.Thread.Title
	tp.Author.Name = otp.Thread.Author.Name
	tp.Author.ID = otp.Thread.Author.ID
	tp.Author.IsLike = otp.Thread.Author.IsLike == 1
	tp.Author.Level = otp.Thread.Author.LevelID
	tp.Author.Portrait = otp.Thread.Author.Portrait

	for i, op := range otp.PostList {
		p := &postList[i]

		p.Pid = op.ID
		p.Floor = op.Floor
		p.PostTime = time.Unix(op.Time, 0)

		p.ContentList = op.Content

		p.Author.ID = op.Author.ID
		p.Author.Name = op.Author.Name
		if level, ok := op.Author.LevelID.(float64); ok {
			p.Author.Level = uint8(level)
		} else {
			p.Author.Level = math.MaxUint8
		}
		if isLike, ok := op.Author.IsLike.(float64); ok {
			p.Author.HasIsLike = true
			p.Author.IsLike = isLike == 1
		} else {
			p.Author.HasIsLike = false
		}
		p.Author.Portrait = op.Author.Portrait
		//p.Thread = &tp
	}

	return &tp, postList, &tpe, nil, nil
}
