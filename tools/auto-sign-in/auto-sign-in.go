package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/BurntSushi/toml"

	//"github.com/purstal/go-tieba-base/http"
	"github.com/purstal/go-tieba-base/logs"
	"github.com/purstal/go-tieba-base/tieba"
	"github.com/purstal/go-tieba-base/tieba/apis"
)

type Config struct {
	PushBullet_AccessToken string
	Tasks                  []Task
}

type Task struct {
	BDUSS    string
	UserName string
}

func main() {

	os.Mkdir("log/", 0644)

	var logfile, _ = os.Create("log/" + time.Now().Format("auto-sign-in-20060102-150405.log"))

	logs.SetDefaultLogger(logs.NewLogger(logs.DebugLevel, os.Stdout, logfile))
	logs.DefaultLogger.LogWithTime = true

	var config_filename string

	if len(os.Args) == 1 {
		config_filename = "config.toml"
	} else {
		config_filename = os.Args[1]
	}

	data, err1 := ioutil.ReadFile(config_filename)

	if err1 != nil {
		logs.Fatal(err1)
		return
	}

	var config Config

	err2 := toml.Unmarshal(data, &config)

	if err2 != nil {
		logs.Fatal(err2)
		return
	}

	//logs.Debug(config)

	var token = config.PushBullet_AccessToken

	var now = time.Now()
	var results []string

	/* -------------------------------------------------------------------------------------------------------------------------------- */
	retryFunc := func(accAndr *postbar.Account, fail_forums []ForumFailed) ([]Forum, []Forum, []ForumFailed, *postbar.PbError) {
		logs.Info(accAndr.ID, "开始重试失败的签到.")
		var forums_retry []Forum
		for _, forum := range fail_forums {
			forums_retry = append(forums_retry, Forum{forum.ForumID, forum.ForumName, nil})
		}
		return signAll(accAndr, forums_retry, time.Second)
	}
	/* -------------------------------------------------------------------------------------------------------------------------------- */

	/* -------------------------------------------------------------------------------------------------------------------------------- */
	finalFunc := func(whom string, beginTime time.Time, succCount, beforeSignedCount int, failForums []ForumFailed) {
		result := makeResult(whom, beginTime, succCount, beforeSignedCount, failForums)
		logs.Info(result)
		results = append(results, result)
	}
	finalFunc_failed := func(whom string, pberr *postbar.PbError) {
		result := fmt.Sprintf(`#%s#
error_code:%d
error_msg:%s
`, whom, pberr.ErrorCode, pberr.ErrorMsg)
		logs.Info(result)
		results = append(results, result)
	}
	/* -------------------------------------------------------------------------------------------------------------------------------- */

	for _, task := range config.Tasks {
		var begin = time.Now()
		accAndr := postbar.NewDefaultAndroidAccount(task.UserName)
		accAndr.BDUSS = task.BDUSS
		accWin8 := postbar.NewDefaultWindows8Account(task.UserName)
		accWin8.BDUSS = task.BDUSS

		//未签到的贴吧(msign),其他贴吧(sign),已签到的贴吧(msign)
		forums_msign, forums_other, forums_signed_msign, pberr1 := getForumList(accAndr, accWin8)
		if pberr1 != nil && pberr1.ErrorCode != 0 {
			finalFunc_failed(task.UserName, pberr1)
			continue
		}

		//成功签到的贴吧(msign),失败签到的贴吧(msign)
		logs.Info(accAndr.ID, "开始普通签到A.")
		succ_msign, fail_msign, pberr2 := mSignAll_fake(accAndr, forums_msign, time.Second)
		if pberr2 != nil && pberr2.ErrorCode != 0 {
			finalFunc_failed(task.UserName, pberr2)
			continue
		} else if len(forums_other) == 0 {
			var succCount = len(succ_msign)
			if len(fail_msign) != 0 {
				succ_retry, signed_retry, fail_retry, pberr := retryFunc(accAndr, fail_msign)
				if pberr != nil && pberr.ErrorCode != 0 {
					finalFunc_failed(task.UserName, pberr)
					continue
				}
				succCount += len(succ_retry) + len(signed_retry)
				fail_msign = fail_retry
			}
			finalFunc(task.UserName, begin, succCount, len(forums_signed_msign), fail_msign)
			continue
		}

		//成功签到的贴吧(sign),失败签到的贴吧(sign),已签到的贴吧(sign)
		logs.Info(accAndr.ID, "开始普通签到B.")
		succ_sign, signed_sign, fail_sign, pberr3 := signAll(accAndr, forums_other, time.Second)
		if pberr3 != nil && pberr3.ErrorCode != 0 {
			finalFunc_failed(task.UserName, pberr3)
			continue
		}

		var fail_all = append(fail_msign, fail_sign...)
		var succCount = len(succ_msign) + len(succ_sign)
		if len(fail_all) != 0 {
			succ_retry, signed_retry, fail_retry, pberr := retryFunc(accAndr, fail_all)
			if pberr != nil && pberr.ErrorCode != 0 {
				finalFunc_failed(task.UserName, pberr)
				continue
			}
			succCount += len(succ_retry) + len(signed_retry)
			fail_all = fail_retry
		}
		finalFunc(task.UserName, begin, succCount, len(forums_signed_msign)+len(signed_sign), fail_all)

	}

	var resultText string = "\n"
	for _, result := range results {
		resultText += result + "\n\n"
	}
	//logs.Debug(resultText)
	for i := 0; i < 10; i++ {
		err := pushNote(token, now.Format("2006-01-02 贴吧签到结果"), resultText)
		if err == nil {
			break
		}
	}

}

func mSignAll_fake(accAndr *postbar.Account, forums []Forum, interval time.Duration) ([]Forum, []ForumFailed, *postbar.PbError) {
	var succ_msign []Forum
	var fail_msign []ForumFailed

	for _, forum := range forums {
		var tbs string
		for {
			var err error
			tbs, err, _ = apis.GetTbs(accAndr)
			if err == nil || tbs == "" {
				break
			}
		}
		ok, info, pberr := sign(accAndr, forum.ForumID, forum.ForumName, tbs, time.Second*10)
		if !ok {
			if pberr != nil && pberr.ErrorCode == 1 {
				return nil, nil, pberr
			}
			logs.Info(accAndr.ID, forum.ForumName, "签到失败,信息:", info)
			fail_msign = append(fail_msign, ForumFailed{forum.ForumID, forum.ForumName, pberr})
		} else {
			logs.Info(accAndr.ID, forum.ForumName, "签到成功,信息:", info)
			succ_msign = append(succ_msign, Forum{forum.ForumID, forum.ForumName, info})
		}
		time.Sleep(interval)
	}

	return succ_msign, fail_msign, nil
}

func signAll(accAndr *postbar.Account, forums []Forum, interval time.Duration) ([]Forum, []Forum, []ForumFailed, *postbar.PbError) {
	var succ_sign, signed_sign []Forum
	var fail_sign []ForumFailed

	for _, forum := range forums {
		yes, info_1, tbs := isSignedIn(accAndr, forum.ForumName)
		if yes {
			logs.Info(accAndr.ID, forum.ForumName, "已经签到,信息:", marshalIgnoreError(info_1))
			signed_sign = append(signed_sign, Forum{forum.ForumID, forum.ForumName, info_1})
			continue
		}
		ok, info, pberr := sign(accAndr, forum.ForumID, forum.ForumName, tbs, interval*time.Duration(10))
		if !ok {
			if pberr != nil && pberr.ErrorCode == 1 {
				return nil, nil, nil, pberr
			}
			logs.Info(accAndr.ID, forum.ForumName, "签到失败,信息:", info)
			fail_sign = append(fail_sign, ForumFailed{forum.ForumID, forum.ForumName, pberr})
		} else {
			logs.Info(accAndr.ID, forum.ForumName, "签到成功,信息:", info)
			succ_sign = append(succ_sign, Forum{forum.ForumID, forum.ForumName, info})
		}
		time.Sleep(interval)
	}
	return succ_sign, signed_sign, fail_sign, nil
}

func marshalIgnoreError(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}

func sign(accAndr *postbar.Account, forumID uint64, forumName, _tbs string, retryInterval time.Duration) (bool, SignInfo, *postbar.PbError) {

	var retryTimes int
	var tbs = _tbs
OUTER:
	for {
		info, err, pberr := apis.Sign(accAndr, forumID, forumName, tbs)
		if err != nil {
			yes, info, _tbs := isSignedIn(accAndr, forumName)
			if yes {
				return true, info, nil
			} else {
				tbs = _tbs
				time.Sleep(retryInterval)
				continue OUTER
			}
		} else {
			if pberr != nil {
				switch pberr.ErrorCode {
				case 1: //没有登陆
					if yes, _ := apis.IsLogin(accAndr.BDUSS); !yes {
						return false, nil, pberr
					} else if retryTimes < 6 {
						retryTimes++
						time.Sleep(retryInterval)
						continue OUTER //安心
					} else {
						return false, SignInfo{
							"重试次数":       retryTimes,
							"error_code": pberr.ErrorCode,
							"error_msg":  pberr.ErrorMsg}, pberr
					}
				case 160002, 199901, 160004: //已经签过,被封禁,不支持
					return true, SignInfo{
						"error_code": pberr.ErrorCode,
						"error_msg":  pberr.ErrorMsg}, pberr
				case 160008: //太快了
					if retryTimes < 6 {
						retryTimes++
						logs.Warn(accAndr.ID, forumName, "签到过快,睡眠后重试:", retryInterval*time.Duration(retryTimes+1))
						time.Sleep(retryInterval * time.Duration(retryTimes+1))
						continue OUTER
					} else {
						return false, SignInfo{
							"重试次数":       retryTimes,
							"error_code": pberr.ErrorCode,
							"error_msg":  pberr.ErrorMsg}, pberr
					}
				default: //鬼知道
					return false, SignInfo{
						"error_code": pberr.ErrorCode,
						"error_msg":  pberr.ErrorMsg}, pberr
				}
			} else {
				return true, SignInfo{
					"累计签到天数": info.TotalSignNum,
					"连续签到天数": info.ContSignNum,
					"当前经验":   info.TotalSignNum,
					"签到时间":   info.SignTime,
					"签到排名":   info.UserSignRank}, nil
			}
		}
	}
}

type Forum struct {
	ForumID   uint64
	ForumName string
	Status    interface{}
}

type ForumFailed struct {
	ForumID   uint64
	ForumName string
	Error     *postbar.PbError
}

func getForumList(accAndr, accWin8 *postbar.Account) (
	forums_msign, forums_other, signed_forums []Forum, pberr *postbar.PbError) {

	var retryTimes int
	var forumCount_1 int = -1
	var canUseMsign bool
	var mSignSet = map[uint64]struct{}{}

	for {
		forumList, err, pberr := apis.GetForumList(accAndr)
		if pberr != nil && pberr.ErrorCode != 0 {
			if retryTimes < 3 {
				retryTimes++
				continue
			} else {
				if pberr.ErrorCode == 1 {
					return nil, nil, nil, pberr
				}
				logs.Error("无法通过 c/f/Forum/getforumlist 获取关注贴吧,将不使用一键签到进行签到(虽然成功了也不用它).")
				break
			}
		}
		if err == nil {
			//logs.Debug(forumList, err, pberr)
			canUseMsign = forumList.CanUse
			forumCount_1 = len(forumList.ForumInfo)
			for _, info := range forumList.ForumInfo {
				mSignSet[info.ForumID] = struct{}{}
				if info.IsSignIn {
					signed_forums = append(signed_forums, Forum{info.ForumID, info.ForumName,
						SignInfo{
							"连续签到天数": info.ContSignNumber,
							"当前经验":   info.UserExp}})
				} else {
					forums_msign = append(forums_msign, Forum{info.ForumID, info.ForumName, nil})
				}
			}
			break
		}

	}

	if forumCount_1 >= 0 && (forumCount_1 < 50 || (forumCount_1 < 200 && canUseMsign)) {
		return forums_msign, nil, signed_forums, nil
	}

	retryTimes = 0
	var forumCount_2 int = -1

	for {
		forumList, err, pberr := apis.GetUserForumLike(accWin8)
		if pberr != nil && pberr.ErrorCode != 0 {
			if retryTimes < 3 {
				retryTimes++
			} else {
				if pberr.ErrorCode == 1 {
					return nil, nil, nil, pberr
				}
				logs.Error("无法通过 c/f/Forum/like 获取关注贴吧.")
				break
			}
		}

		if err == nil {
			for _, info := range forumList {
				if _, exist := mSignSet[info.ID]; !exist {
					forums_other = append(forums_other, Forum{info.ID, info.ForumName, nil})
				}
			}
			break
		}
	}

	if forumCount_2 >= 0 && forumCount_2 < 200 {
		return forums_msign, forums_other, signed_forums, nil
	}

	//反正我也没有签到超过200个贴吧的需求,
	//就不用 http://tieba.baidu.com/f/like/mylike 来获取余下的贴吧了!

	return forums_msign, forums_other, signed_forums, nil
}

type SignInfo map[string]interface{}

func isSignedIn(accAndr *postbar.Account, forumName string) (bool, SignInfo, string) {
	var st struct {
		Anti struct {
			TBS string `json:"tbs"`
		} `json:"anti"`
		Forum struct {
			CurScore   int `json:"cur_score,string"`
			SignInInfo struct {
				UserInfo struct {
					IsSignIn     string `json:"is_sign_in"`
					ContSignNum  int    `json:"cont_sign_num,string"`
					TotlaSignNum int    `json:"cout_total_sing_num,string"` //什么鬼!
					SignTime     int64  `json:"sign_time,string"`
					UserSignRank int    `json:"user_sign_rank,string"`
				} `json:"user_info"`
			} `json:"sign_in_info"`
		} `json:"Forum"`
	}
	for {
		data, err := apis.RGetForum(accAndr, forumName, 0, 1)
		if err == nil && len(data) != 0 {
			err := json.Unmarshal(data, &st)
			if err == nil {
				return st.Forum.SignInInfo.UserInfo.IsSignIn == "1", SignInfo{
					"累计签到天数": st.Forum.SignInInfo.UserInfo.TotlaSignNum,
					"连续签到天数": st.Forum.SignInInfo.UserInfo.ContSignNum,
					"当前经验":   st.Forum.CurScore,
					"签到时间":   time.Unix(st.Forum.SignInInfo.UserInfo.SignTime, 0),
					"签到排名":   st.Forum.SignInInfo.UserInfo.UserSignRank}, st.Anti.TBS
			}
		}
	}
}

/*
 curl --header 'Authorization: Bearer <your_access_token_here>'
 -X POST https://api.pushbullet.com/v2/pushes
 --header 'Content-Type: application/json'
 --data-binary '{"type": "note", "title": "Note Title", "body": "Note Body"}'
*/
func pushNote(token string, title, body string) error {

	var data = struct {
		Type  string `json:"type"`
		Title string `json:"title"`
		Body  string `json:"body"`
	}{"note", title, body}

	data_json, _ := json.Marshal(data)

	var request, _ = http.NewRequest("POST", "https://api.pushbullet.com/v2/pushes", bytes.NewReader(data_json))
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(request)
	if err == nil {
		request.Body.Close()
		a, _ := ioutil.ReadAll(resp.Body)
		logs.Debug(string(a))
	} else {
		logs.Debug(err)
	}

	return err

}

func makeResult(whom string, beginTime time.Time, succCount, beforeSignedCount int, failForums []ForumFailed) string {
	var body, fail_info string

	failMap := map[int]*struct {
		Msg    string
		Forums []string
	}{}

	for _, forum := range failForums {
		if st, exist := failMap[forum.Error.ErrorCode]; exist {
			st.Forums = append(st.Forums, forum.ForumName)
		} else {
			failMap[forum.Error.ErrorCode] = &struct {
				Msg    string
				Forums []string
			}{forum.Error.ErrorMsg, []string{forum.ForumName}}
		}
	}

	for code, st := range failMap {
		fail_info += fmt.Sprintf("--%d(%s)%s\n", code, st.Msg, st.Forums)
	}

	body = fmt.Sprintf(`#%s# 签到成功
开始签到时间:%s
成功签到数:%d
早前已签到数:%d
失败签到:%d
%s`, whom, beginTime.Format("2006-01-02 15:04:05"), succCount, beforeSignedCount, len(failForums), fail_info)
	//↑不换行是因为已经包含换行符

	return body
}
