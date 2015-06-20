package forum

import (
	//"encoding/json"
	"time"

	"github.com/purstal/pbtools/modules/postbar"
)

func GetForumStruct(
	acc *postbar.Account, kw string, rn,
	pn int) (*ForumPage, []*ForumPageThread, *ForumPageExtra, error, *postbar.PbError) {

	ofp, err, pberr := GetOriginalForumStruct(acc, kw, rn, pn)

	if err != nil {
		return nil, nil, nil, err, nil
	}

	var fpe ForumPageExtra
	fpe.IsLogin = ofp.User.IsLogin == 1
	fpe.ServerTime = time.Unix(ofp.Time, 0)
	fpe.LogID = ofp.LogID
	if pberr != nil {
		return nil, nil, &fpe, nil, pberr
	}

	var fp ForumPage
	var ThreadList = make([]*ForumPageThread, len(ofp.ThreadList))

	fp.Fid = ofp.Forum.ID
	fp.ForumName = ofp.Forum.Name

	for i, ot := range ofp.ThreadList {
		t := &ForumPageThread{} //&ThreadList[i]
		ThreadList[i] = t
		t.Tid = ot.Tid
		t.Title = ot.Title
		t.ReplyNum = ot.ReplyNum

		t.LastReplyTime = time.Unix(ot.LastTimeInt, 0)
		t.IsTop, t.IsGood = ot.IsTop == 1, ot.IsGood == 1

		t.Author.Name = ot.Author.Name
		t.Author.ID = ot.Author.ID
		t.Author.Portrait = ot.Author.Portrait
		t.LastReplyer.Name = ot.LastReplyer.Name
		t.LastReplyer.ID = ot.LastReplyer.ID
		t.MediaList = ot.Media
		//t.MediaList, _ = json.Marshal(ot.Media)
		/*
			if len(ot.Abstract) != 0 {
				t.AbstractText = ot.Abstract[0].Text
			} else {
				t.AbstractText = ""
			}
		*/
		t.Abstract = ot.Abstract
		//t.Abstract, _ = json.Marshal(ot.Abstract)

		//t.Forum = &fp

	}

	return &fp, ThreadList, &fpe, nil, nil
}
