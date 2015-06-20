package thread

import (
	"math"
	"time"

	"github.com/purstal/pbtools/modules/postbar"
)

func GetThreadStruct(acc *postbar.Account, tid uint64, mark bool, pid uint64, pn, rn int,
	withFloor, seeLz, r bool) (*ThreadPage, []ThreadPagePost,
	*ThreadPageExtra, error, *postbar.PbError) {

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
