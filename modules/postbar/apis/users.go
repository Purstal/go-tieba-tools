package apis

import (
	"encoding/json"
	"strconv"

	"github.com/purstal/pbtools/modules/http"
	"github.com/purstal/pbtools/modules/postbar/accounts"
)

type ForumLikePageForumList struct {
	NonGconForum []ForumLikePageForum `json:"non-gconforum"`
	GconForum    []ForumLikePageForum `json:"gconforum"`
}

type ForumLikePageForum struct {
	ID           uint64 `json:"id,string"`
	ForumName    string `json:"name"`
	Level        int    `json:"level_id,string"`
	CurrentScore int    `json:"cur_score,string"`
}

//无论隐藏与否
func GetUserForumLike(acc *accounts.Account, uid uint64) (*ForumLikePageForumList, error) {

	var parameters http.Parameters

	parameters.Add("uid", strconv.FormatUint(uid, 10))

	accounts.ProcessParams(&parameters, acc)

	resp, err := http.Post("http://c.tieba.baidu.com/c/f/forum/like", parameters)
	if err != nil {
		return nil, err
	}

	var data struct {
		ForumList ForumLikePageForumList `json:"forum_list"`
	}
	err2 := json.Unmarshal(resp, &data)

	return &data.ForumList, err2

}

func _GetUserProfile(acc *accounts.Account, uid uint64) {
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

func GetUserInfo(acc *accounts.Account, uid uint64) (*UserInfo, error) {
	var parameters http.Parameters

	parameters.Add("uid", strconv.FormatUint(uid, 10))
	accounts.ProcessParams(&parameters, acc)

	resp, err := http.Post("http://c.tieba.baidu.com/c/u/user/getuserinfo", parameters)

	if err != nil {
		return nil, err
	}

	var userInfo UserInfo

	err2 := json.Unmarshal(resp, &userInfo)

	return &userInfo, err2
}
