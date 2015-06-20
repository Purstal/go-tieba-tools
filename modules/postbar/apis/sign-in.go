package apis

import (
	"encoding/json"
	//"fmt"

	"github.com/purstal/pbtools/modules/http"
	"github.com/purstal/pbtools/modules/postbar"
)

type ForumList struct {
	CanUse    bool
	ForumInfo []struct {
		ForumID   uint64
		ForumName string
		IsSignIn  bool
		UserLevel int
	}
}

func GetForumList(accAndr *postbar.Account) (*ForumList, error, *postbar.PbError) {

	accAndr.ClientVersion = "6.6.6"

	var _forumList struct {
		CanUse    string `json:"can_use"`
		ErrorCode int    `json:"error_code,string"`
		ErrorMsg  string `json:"error_msg"`
		ForumInfo []struct {
			ForumID   uint64 `json:"forum_id,string"`
			ForumName string `json:"forum_name"`
			IsSignIn  string `json:"is_sign_in"`
			UserLevel int    `json:"user_level,string"`
		} `json:"forum_info"`
	}

	var parameters http.Parameters
	postbar.ProcessParams(&parameters, accAndr)

	resp, err := http.Post("http://c.tieba.baidu.com/c/f/forum/getforumlist", parameters)

	if err != nil {
		return nil, err, nil
	}

	json.Unmarshal(resp, &_forumList)

	if _forumList.ErrorCode != 0 {
		return nil, nil, postbar.NewPbError(_forumList.ErrorCode, _forumList.ErrorMsg)
	}

	var forumList ForumList

	forumList.CanUse = _forumList.CanUse == "1"

	for _, info := range _forumList.ForumInfo {
		forumList.ForumInfo = append(forumList.ForumInfo, struct {
			ForumID   uint64
			ForumName string
			IsSignIn  bool
			UserLevel int
		}{info.ForumID, info.ForumName, info.IsSignIn == "1", info.UserLevel})
	}

	return &forumList, nil, nil

}
