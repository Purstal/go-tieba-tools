package apis

import (
	"encoding/json"
	"strconv"

	//"github.com/purstal/pbtools/misc"

	"github.com/purstal/pbtools/modules/http"
	"github.com/purstal/pbtools/modules/postbar"
)

func DeletePost(acc *postbar.Account, pid uint64) (error, *postbar.PbError) {
	//无论是主题/回复,亦或是楼中楼,只要有pid都可以用这个删..
	var parameters http.Parameters

	parameters.Add("pid", strconv.FormatUint(pid, 10))
	//↑楼中楼要spid,如果用pid会删掉整个楼层,因为pid是楼层的..

	tbs, err, pberr := GetTbs(acc)
	if err != nil {
		return err, nil
	} else if pberr != nil {
		return nil, pberr
	}
	parameters.Add("tbs", tbs)

	postbar.ProcessParams(&parameters, acc)

	resp, err := http.Post("http://c.tieba.baidu.com/c/c/bawu/delpost", parameters)
	if err != nil {
		return err, nil
	}
	return postbar.ExtractError(resp)

}

func DeleteThread(acc *postbar.Account, tid uint64) (error, *postbar.PbError) {
	var parameters http.Parameters

	tbs, err, pberr := GetTbs(acc)
	if err != nil {
		return err, nil
	} else if pberr != nil {
		return nil, pberr
	}
	parameters.Add("tbs", tbs)

	parameters.Add("z", strconv.FormatUint(tid, 10))

	postbar.ProcessParams(&parameters, acc)
	resp, err := http.Post("http://c.tieba.baidu.com/c/c/bawu/delthread", parameters)

	if err != nil {
		return err, nil
	}
	return postbar.ExtractError(resp)
}

//pid得(dei)超准..
func BlockIDWeb(BDUSS string,
	forumID uint64, userName string, pid uint64, day int,
	reason string) (error, *postbar.PbError) {

	var parameters http.Parameters
	parameters.Add("day", strconv.Itoa(day))
	parameters.Add("fid", strconv.FormatUint(forumID, 10))

	tbs, err := GetTbsWeb(BDUSS)
	if err != nil {
		return err, nil
	}
	parameters.Add("tbs", tbs)

	parameters.Add("ie", "gbk")
	parameters.Add("user_name[]", userName)
	parameters.Add("pid[]", strconv.FormatUint(pid, 10))
	parameters.Add("reason", reason)

	var cookies http.Cookies
	cookies.Add("BDUSS", BDUSS)

	//println(parameters.Encode())

	resp, err := http.Get("http://tieba.baidu.com/pmc/blockid", parameters, cookies)

	if err != nil {
		return err, nil
	}
	var x struct {
		ErrorCode int    `json:"errno"`
		ErrorMsg  string `json:"errmsg"`
	}

	json.Unmarshal(resp, &x)

	return nil, postbar.NewPbError(x.ErrorCode, x.ErrorMsg)
}

func CommitPrison(account *postbar.Account,
	forumName string, forumID uint64, userName string, threadID, postID uint64,
	day int, reason string) (error, *postbar.PbError) {
	var parameters http.Parameters
	parameters.Add("day", strconv.Itoa(day))
	parameters.Add("fid", strconv.FormatUint(forumID, 10))
	parameters.Add("ntn", "banid") //"banip"会出错
	parameters.Add("post_id", strconv.FormatUint(postID, 10))
	parameters.Add("reason", reason)
	tbs, err := GetTbsWeb(account.BDUSS)
	if err != nil {
		return err, nil
	}
	parameters.Add("tbs", tbs)
	parameters.Add("un", userName)
	parameters.Add("word", forumName)
	parameters.Add("z", strconv.FormatUint(threadID, 10))
	postbar.ProcessParams(&parameters, account)

	resp, err := http.Post("http://c.tieba.baidu.com/c/c/bawu/commitprison", parameters)
	if err != nil {
		return err, nil
	}
	return postbar.ExtractError(resp)
}
