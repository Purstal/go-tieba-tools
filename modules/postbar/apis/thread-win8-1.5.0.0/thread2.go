package thread

import (
	"time"

	"github.com/purstal/pbtools/modules/postbar"
)

type Forum struct {
	ID   uint64
	Name string
}

type Thread struct {
	Forum Forum

	Page struct {
		CurrentPage int
		TotalPage   int
	}

	PostList []ThreadPagePost
	Thread   ThreadPage

	Time time.Time

	//Original interface{}
}

func GetThread2(accWin8 *postbar.Account, tid uint64, mark bool, pid uint64, pn, rn int,
	withFloor, seeLz, r bool) (*Thread, error, *postbar.PbError) {

	_thread, err, pberr := GetOriginalThreadStruct(accWin8, tid, mark, pid, pn, rn, withFloor, seeLz, r)

	if err != nil {
		return nil, err, nil
	}

	if pberr != nil && pberr.ErrorCode != 0 {
		return nil, nil, pberr
	}

	var thread Thread

	thread.Forum.ID = _thread.Forum.ID
	thread.Forum.Name = _thread.Forum.Name

	thread.Page.CurrentPage = _thread.Page.CurrentPage
	thread.Page.TotalPage = _thread.Page.TotalPage

	thread.PostList = make([]ThreadPagePost, len(_thread.PostList))

	for i, _post := range _thread.PostList {
		post := &thread.PostList[i]

		post.Pid = _post.ID
		post.Floor = _post.Floor
		post.PostTime = time.Unix(_post.Time, 0)

		post.ContentList = _post.Content

		post.Author.ID = _post.Author.ID
		post.Author.Name = _post.Author.Name

		if level, ok := _post.Author.LevelID.(float64); ok {
			post.Author.Level = uint8(level)
		}
		if isLike, ok := _post.Author.IsLike.(float64); ok {
			post.Author.HasIsLike = true
			post.Author.IsLike = isLike == 1
		} else {
			post.Author.HasIsLike = false
		}

		post.Author.Portrait = _post.Author.Portrait
	}

	thread.Thread.Tid = _thread.Thread.ID
	thread.Thread.Title = _thread.Thread.Title
	thread.Thread.Author.Name = _thread.Thread.Author.Name
	thread.Thread.Author.ID = _thread.Thread.Author.ID
	thread.Thread.Author.IsLike = _thread.Thread.Author.IsLike == 1
	thread.Thread.Author.Level = _thread.Thread.Author.LevelID
	thread.Thread.Author.Portrait = _thread.Thread.Author.Portrait

	thread.Time = time.Unix(_thread.Time, 0)

	return &thread, nil, nil

}
