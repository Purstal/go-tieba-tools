package floor

import (
	"encoding/json"
	"os"
	"strconv"
	"time"

	"github.com/purstal/pbtools/modules/pberrors"
	"github.com/purstal/pbtools/modules/postbar"
	"github.com/purstal/pbtools/modules/postbar/apis"
)

type OriginalFloorStruct struct {
	CommentList []struct {
		ID      uint64        `json:"id,string"`
		Content []interface{} `json:"content"`
		Time    int64         `json:"time,string"`
		Author  struct {
			ID       uint64 `json:"id,string"`
			Name     string `json:"name"`
			LevelID  uint8  `json:"level_id,string"`
			Portrait string `json:"portrait"`
			//is_like无效
		}
	} `json:"subpost_list"`
	Post struct {
		ID      uint64        `json:"id,string"`
		Floor   int           `json:"floor,string"`
		Time    int64         `json:"time,string"`
		Content []interface{} `json:"content"`
		Author  struct {
			ID       uint64 `json:"id,string"`
			Name     string `json:"name"`
			LevelID  uint8  `json:"level_id,string"`
			IsLike   int    `json:"is_like,string"`
			Portrait string `json:"portrait"`
		} `json:"author"`
	} `json:"post"`
	Page struct {
		TotalPage   int    `json:"total_page,string"`
		TotalCount  string `json:"total_count"`
		CurrentPage int    `json:"current_page,string"`
		PageSize    int    `json:"page_size,string"`
	} `json:"page"`
	Thread struct {
		ID     uint64 `json:"id,string"`
		Title  string `json:"title"`
		Author struct {
			ID       uint64 `json:"id,string"`
			Name     string `json:"name"`
			LevelID  uint8  `json:"level_id,string"`
			IsLike   int    `json:"is_like,string"`
			Portrait string `json:"portrait"`
		} `json:"author"`
	} `json:"thread"`
	Time      int64  `json:"time"`
	ErrorCode int    `json:"error_code,string"`
	ErrorMsg  string `json:"error_msg"`
}

func GetOriginalFloorStruct(acc *postbar.Account, tid uint64,
	isComment bool, id uint64, pn int) (*OriginalFloorStruct, error, *pberrors.PbError) {
	resp, err := apis.RGetFloor(acc, tid, isComment, id, pn)

	if err != nil {
		return nil, err, nil
	}
	var x OriginalFloorStruct
	err2 := json.Unmarshal(resp, &x)
	if err2 != nil {
		os.MkdirAll("err/wrongJSON", 0744)
		f, _ := os.Create("err/wrongJSON/" + strconv.FormatInt(time.Now().Unix(), 16) + ".json")
		if err != nil {
			f.Write(resp)
			f.Close()
		}
		return nil, err2, nil
	}

	if x.ErrorCode != 0 {
		return &x, nil, pberrors.NewPbError(x.ErrorCode, x.ErrorMsg)
	}
	return &x, nil, nil

}
