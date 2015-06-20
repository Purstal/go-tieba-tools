package apis

import (
	"encoding/json"

	"github.com/purstal/pbtools/modules/http"
	"github.com/purstal/pbtools/modules/postbar"
)

func RSearchForum(query string) ([]byte, error) {
	var parameters http.Parameters
	parameters.Add("query", query)
	postbar.AddSignature(&parameters)
	return http.Post(`http://c.tieba.baidu.com/c/f/forum/search`, parameters)
}

type ForumSearchResult struct {
	ForumID   uint64 `json:"forum_id"`
	ForumName string `json:"forum_name"`
}

func SearchForum(query string) ([]ForumSearchResult, error, *postbar.PbError) {
	resp, err := RSearchForum(query)
	if err != nil {
		return nil, err, nil
	}

	var forumSearchResults struct {
		ForumList []ForumSearchResult `json:"forum_list"`
		ErrorCode int                 `json:"error_code,string"`
		ErrorMsg  string              `json:"error_msg"`
	}

	json.Unmarshal(resp, &forumSearchResults)

	if forumSearchResults.ErrorCode == 110003 {
		return nil, nil, nil
	} else if forumSearchResults.ErrorCode != 0 {
		return nil, nil, postbar.NewPbError(forumSearchResults.ErrorCode, forumSearchResults.ErrorMsg)
	}

	return forumSearchResults.ForumList, nil, nil

}
