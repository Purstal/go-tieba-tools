package apis

import (
	"encoding/json"
	"strconv"

	"github.com/purstal/pbtools/modules/http"
	"github.com/purstal/pbtools/modules/postbar"
)

/*
type ForumLikePageForumList struct {
	NonGconForum []ForumLikePageForum `json:"non-gconforum"`
	GconForum    []ForumLikePageForum `json:"gconforum"`
}
*/

type ForumLikePageForum struct {
	ID           uint64 `json:"id,string"`
	ForumName    string `json:"name"`
	Level        int    `json:"level_id,string"`
	CurrentScore int    `json:"cur_score,string"`
}

//无论隐藏与否 //只能取前200个贴吧
func GetUserForumLike(accWin8 *postbar.Account /*, uid uint64*/) ([]ForumLikePageForum, error, *postbar.PbError) {

	var parameters http.Parameters

	//parameters.Add("uid", strconv.FormatUint(uid, 10))

	postbar.ProcessParams(&parameters, accWin8)

	resp, err := http.Post("http://c.tieba.baidu.com/c/f/forum/like", parameters)
	if err != nil {
		return nil, err, nil
	}

	var data struct {
		ForumList []ForumLikePageForum `json:"forum_list"`
		ErrorCode int                  `json:"error_code,string"`
		ErrorMsg  string               `json:"error_msg"`
	}

	err2 := json.Unmarshal(resp, &data)

	if err2 != nil {
		return nil, err2, nil
	}

	if data.ErrorCode != 0 {
		return nil, nil, postbar.NewPbError(data.ErrorCode, data.ErrorMsg)
	}

	return data.ForumList, nil, nil

}

func _GetUserProfile(acc *postbar.Account, uid uint64) {
	/*
		var parameters http.Parameters

		parameters.Add("uid", strconv.FormatUint(uid, 10))

		ProcessParams(&parameters, acc)

		resp, err := http.Post("http://c.tieba.baidu.com/c/u/user/profile", parameters)
	*/
}

type UserInfo struct {
	User struct {
		ID       uint64 `json:"id,string"`
		Name     string `json:"name"`
		Portrait string `json:"portrait"`
		Sex      int    `json:"sex,string"`
	} `json:"user"`

	ErrorCode interface{} `json:"error_code"`
}

func GetUserInfo(acc *postbar.Account, uid uint64) (*UserInfo, error) {
	var parameters http.Parameters

	parameters.Add("uid", strconv.FormatUint(uid, 10))
	postbar.ProcessParams(&parameters, acc)

	resp, err := http.Post("http://c.tieba.baidu.com/c/u/user/getuserinfo", parameters)

	if err != nil {
		return nil, err
	}

	var userInfo UserInfo

	err2 := json.Unmarshal(resp, &userInfo)

	return &userInfo, err2
}
