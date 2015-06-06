package thread

import (
	"encoding/json"
	"os"

	"github.com/purstal/pbtools/modules/pberrors"
	"github.com/purstal/pbtools/modules/postbar"
)

type OriginalThreadStruct struct {
	Forum struct {
		ID   uint64 `json:"id,string"`
		Name string
	}

	Page struct {
		CurrentPage int `json:"current_page,string"`
		TotalPage   int `json:"total_page,string"`
	}

	PostList []struct {
		ID    uint64 `json:"id,string"`
		Floor int    `json:"floor,string"`
		Time  int64  `json:"time,string"`

		Content []interface{} `json:"content"`
		Author  struct {
			ID       uint64      `json:"id,string"`
			Name     string      `json:"name"`
			LevelID  interface{} `json:"level_id"`
			IsLike   interface{} `json:"is_like"`
			Portrait string      `json:"portrait"`
		} `json:"author"`
	} `json:"post_list"`

	Thread struct {
		ID         uint64 `json:"id,string"`
		CreateTime uint64 `json:"create_time,string"`
		Title      string `json:"title"`
		Author     struct {
			ID       uint64 `json:"id,string"`
			Name     string `json:"name"`
			LevelID  uint8  `json:"level_id,string"`
			IsLike   int    `json:"is_like,string"`
			Portrait string `json:"portrait"`
		} `json:"author"`
	}

	Time int64 `json:"time"`

	ErrorCode int    `json:"error_code,string"`
	ErrorMsg  string `json:"error_msg"`
}

func GetOriginalThreadStruct(acc *postbar.Account, tid uint64, mark bool, pid uint64, pn, rn int,
	withFloor, seeLz, r bool) (*OriginalThreadStruct, error, *pberrors.PbError) {
	resp, err := RGetThread(acc, tid, mark, pid, pn, rn, withFloor, seeLz, r)

	if err != nil {
		return nil, err, nil
	}
	var x OriginalThreadStruct
	err2 := json.Unmarshal(resp, &x)
	if err2 != nil {
		f, _ := os.Create("err/xxx.json")
		f.Write(resp)
		f.Close()
		return nil, err2, nil
	}
	if x.ErrorCode != 0 {
		return &x, nil, pberrors.NewPbError(x.ErrorCode, x.ErrorMsg)
	}
	return &x, nil, nil

}
