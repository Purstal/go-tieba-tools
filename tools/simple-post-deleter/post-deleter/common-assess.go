package post_deleter

import (
	"fmt"
	"math"
	"regexp"

	"github.com/purstal/pbtools/modules/postbar"
	postfinder "github.com/purstal/pbtools/tools-core/post-finder"
)

func (d *PostDeleter) CommonAssess(from string, account *postbar.Account, post postbar.IPost, tid uint64) postfinder.Control {

	_, uid := post.PGetAuthor().AGetID()
	pid := post.PGetPid()

	if _, exist := d.Records.WaterThread_Tids[tid]; exist {
		d.Logger.Debug(MakePrefix(nil, tid, pid, 0, uid), "水楼的贴子应该来不到这里,但是不知道为什么来了.")
		return postfinder.Finish //防止水楼回复被删
	} else if _, exist := d.Tid_Whitelist.KeyWords()[tid]; exist {
		d.Logger.Debug(MakePrefix(nil, tid, pid, 0, uid), "白名单内的贴子应该来不到这里,但是不知道为什么来了.")
		return postfinder.Finish
	} else if InStringSet(d.BawuList.KeyWords(), post.PGetAuthor().AGetName()) ||
		InStringSet(d.UserName_Whitelist.KeyWords(), post.PGetAuthor().AGetName()) {
		d.Logger.Debug(MakePrefix(nil, tid, pid, 0, uid), "白名单内的用户/吧务应该来不到这里,但是不知道为什么来了.")
		return postfinder.Finish
	}

	text := ExtractText(post.PGetContentList())

	var deleteReason, banReason []string

	if matchedExp := MatchAny(text, d.Content_RxKw.KeyWords()); matchedExp != nil {
		if matchedExp.BanFlag {
			banReason = append(banReason, fmt.Sprint("内容匹配关键词:", matchedExp))
		}
		deleteReason = append(deleteReason, fmt.Sprint("内容匹配关键词:", matchedExp))
	} else if math.Mod(float64(len(text)), 15.0) == 0 {
		if match, _ := regexp.MatchString("[1十拾⑩①][5五伍⑤]字", text); match {
			deleteReason = append(deleteReason, fmt.Sprint("标准十五字"))
		}
	}
	if matchedExp := MatchAny(post.PGetAuthor().AGetName(), d.UserName_RxKw.KeyWords()); matchedExp != nil {
		if matchedExp.BanFlag {
			banReason = append(banReason, fmt.Sprint("用户名匹配关键词:", matchedExp))
		}
		deleteReason = append(deleteReason, fmt.Sprint("用户名匹配关键词:", matchedExp))
	}

	if len(deleteReason) != 0 {
		if len(deleteReason) == 1 {
			d.DeletePost(from, account, tid, pid, 0, uid, deleteReason[0])
		} else {
			d.DeletePost(from, account, tid, pid, 0, uid, fmt.Sprint(deleteReason))
		}
		if len(banReason) == 0 {
			return postfinder.Finish
		}
	}

	if len(banReason) != 0 {
		if len(banReason) == 1 {
			d.BanID(from, account.BDUSS, post.PGetAuthor().AGetName(),
				d.ForumID, tid, pid, uid, 1, banReason[0], "null")
		} else {
			d.BanID(from, account.BDUSS, post.PGetAuthor().AGetName(),
				d.ForumID, tid, pid, uid, 1, fmt.Sprint(banReason), "null")
		}
		return postfinder.Finish
	}

	return postfinder.Continue
}
