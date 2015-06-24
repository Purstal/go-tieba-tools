package apis

import (
	"encoding/json"
	//"fmt"
	"strconv"
	"time"

	"github.com/purstal/pbtools/modules/http"
	"github.com/purstal/pbtools/modules/postbar"
)

type ForumList struct {
	CanUse    bool
	ForumInfo []struct {
		ForumID   uint64
		ForumName string
		IsSignIn  bool
		UserLevel,
		ContSignNumber,
		UserExp int
	}
}

func GetForumList(accAndr *postbar.Account) (*ForumList, error, *postbar.PbError) {

	accAndr.ClientVersion = "6.6.6"

	var _forumList struct {
		CanUse    string `json:"can_use"`
		ErrorCode int    `json:"error_code,string"`
		ErrorMsg  string `json:"error_msg"`
		ForumInfo []struct {
			ForumID        uint64 `json:"forum_id,string"`
			ForumName      string `json:"forum_name"`
			IsSignIn       string `json:"is_sign_in"`
			UserLevel      int    `json:"user_level,string"`
			ContSignNumber int    `json:"cont_sign_num,string"`
			UserExp        int    `json:"user_exp,string"`
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
			UserLevel,
			ContSignNumber,
			UserExp int
		}{info.ForumID, info.ForumName, info.IsSignIn == "1", info.UserLevel, info.ContSignNumber, info.UserExp})
	}

	return &forumList, nil, nil
}

//不知道能不能用
func RMSgin(accAndr *postbar.Account, forum_ids []uint64) ([]byte, error) {
	var tbs string
	for {
		var err error
		tbs, err, _ = GetTbs(accAndr)
		if err == nil {
			break
		}
	}

	var forum_ids_str string
	for _, forum_id := range forum_ids {
		forum_ids_str = forum_ids_str +
			strconv.FormatUint(forum_id, 10) + "%2C"
	}

	var parameters http.Parameters
	parameters.Add("forum_ids", forum_ids_str)
	parameters.Add("tbs", tbs)

	postbar.ProcessParams(&parameters, accAndr)

	return http.Post("http://c.tieba.baidu.com/c/c/forum/msign", parameters)
}

type SignInfo struct {
	IsSignIn     bool
	UserSignRank int
	SignTime     time.Time
	ContSignNum  int
	TotalSignNum int
}

func Sign(accAndr *postbar.Account, fid uint64, kw, tbs string) (*SignInfo, error, *postbar.PbError) {
	var parameters http.Parameters
	parameters.Add("fid", strconv.FormatUint(fid, 10))
	parameters.Add("kw", kw)
	parameters.Add("tbs", tbs)
	postbar.ProcessParams(&parameters, accAndr)
	var resp []byte

	if _resp, err := http.Post("http://c.tieba.baidu.com/c/c/forum/sign", parameters); err != nil {
		return nil, err, nil
	} else {
		resp = _resp
	}

	var signInfo struct {
		ErrorCode int    `json:"error_code,string"`
		ErrorMsg  string `json:"error_msg"`
		UserInfo  struct {
			IsSignIn     string
			UserSignRank int   `json:"user_sign_rank,string"`
			SignTime     int64 `json:"sign_time,string"`
			ContSignNum  int   `json:"cont_sign_num,string"`
			TotalSignNum int   `json:"total_sign_num,string"`
		} `json:"user_info"`
	}

	if err := json.Unmarshal(resp, &signInfo); err != nil {
		return nil, err, nil
	}

	if signInfo.ErrorCode != 0 {
		return nil, nil, postbar.NewPbError(signInfo.ErrorCode, signInfo.ErrorMsg)
		/*
			根据https://github.com/kookxiang/Tieba_Sign/blob/master/system/function/sign.php 以及我自己的判断

			340010,160002:已经签过
			160004:不支持签到(已经没了吧...)
			160003:零点
			160008:快快哒
			199901:被封了,其实成功了
		*/
	}

	return &SignInfo{
			IsSignIn:     signInfo.UserInfo.IsSignIn == "1",
			UserSignRank: signInfo.UserInfo.UserSignRank,
			SignTime:     time.Unix(signInfo.UserInfo.SignTime, 0),
			ContSignNum:  signInfo.UserInfo.ContSignNum,
			TotalSignNum: signInfo.UserInfo.TotalSignNum},
		nil, nil

}
