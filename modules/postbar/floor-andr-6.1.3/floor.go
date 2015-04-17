package floor

import (
	"strconv"
	"time"

	"github.com/purstal/pbtools/modules/http"
	"github.com/purstal/pbtools/modules/pberrors"
	"github.com/purstal/pbtools/modules/postbar/accounts"
	"github.com/purstal/pbtools/modules/postbar/thread-win8-1.5.0.0"
)

func GetFloorJson(acc *accounts.Account, kz uint64,
	isComment bool, id uint64, pn int) ([]byte, error) {
	var parameters http.Parameters
	parameters.Add("from", "tieba")
	parameters.Add("kz", strconv.FormatUint(kz, 10))
	if isComment {
		parameters.Add("spid", strconv.FormatUint(id, 10))
	} else {
		parameters.Add("pid", strconv.FormatUint(id, 10))
	}
	if pn != 0 {
		parameters.Add("pn", strconv.Itoa(pn))
	}

	accounts.ProcessParams(&parameters, acc)

	resp, err := http.Post(`http://c.tieba.baidu.com/c/f/pb/floor`, parameters)
	return resp, err
}

func GetFloorStruct(acc *accounts.Account, kz uint64,
	isComment bool, id uint64, pn int) (*thread.ThreadPage,
	*thread.ThreadPagePost, []FloorPageComment,
	*FloorPageExtra, error, *pberrors.PbError) {

	ofp, err, pberr := GetOriginalFloorStruct(acc, kz, isComment, id, pn)

	if err != nil {
		return nil, nil, nil, nil, err, nil
	}

	var fpe FloorPageExtra
	fpe.ServerTime = time.Unix(ofp.Time, 0)
	if pberr != nil {
		return nil, nil, nil, &fpe, nil, pberr
	}

	var commentPage []FloorPageComment = make([]FloorPageComment, len(ofp.CommentList))
	var tp thread.ThreadPage
	var tpp thread.ThreadPagePost

	tp.Tid = ofp.Thread.ID
	tp.Title = ofp.Thread.Title
	tp.Author.Name = ofp.Thread.Author.Name
	tp.Author.ID = ofp.Thread.Author.ID
	tp.Author.IsLike = ofp.Thread.Author.IsLike == 1
	tp.Author.Level = ofp.Thread.Author.LevelID
	tp.Author.Portrait = ofp.Thread.Author.Portrait

	tpp.Pid = ofp.Post.ID
	tpp.Floor = ofp.Post.Floor
	tpp.PostTime = time.Unix(ofp.Post.Time, 0)
	tpp.ContentList = ofp.Post.Content
	tpp.Author.ID = ofp.Post.Author.ID
	tpp.Author.Name = ofp.Post.Author.Name
	tpp.Author.Level = ofp.Post.Author.LevelID
	tpp.Author.IsLike = ofp.Post.Author.IsLike == 1
	tpp.Author.Portrait = ofp.Post.Author.Portrait

	for i, oc := range ofp.CommentList {
		c := &commentPage[i]

		c.Spid = oc.ID
		c.PostTime = time.Unix(oc.Time, 0)

		c.ContentList = oc.Content

		c.Author.ID = oc.Author.ID
		c.Author.Name = oc.Author.Name
		c.Author.Level = oc.Author.LevelID
		c.Author.Portrait = oc.Author.Portrait
	}

	return &tp, &tpp, commentPage, &fpe, nil, nil

}
