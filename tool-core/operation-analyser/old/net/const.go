package net

import (
	"time"
)

const (
	HOST_ADDRESS = "http://c.tieba.baidu.com/c/"

	AddStore = 0x22

	FORUM_LIKE       = 12
	FORUM_SIGN       = 0x20
	FORUM_UNFAVOLIKE = 0x11
	FORUM_UNLIKE     = 15

	PARAM_ANONYMOUS      = "anonymous"
	PARAM_BACK           = "back"
	PARAM_BDUSS          = "BDUSS"
	PARAM_CLIENT_ID      = "_client_id"
	PARAM_CLIENT_TYPE    = "_client_type"
	PARAM_CLIENT_VERSION = "_client_version"
	PARAM_CONTENT        = "content"
	PARAM_FID            = "fid"
	PARAM_FLOOR_NUM      = "floor_num"
	PARAM_FRS_RN         = 30
	PARAM_ISGOOD         = "is_good"
	PARAM_ISPHONE        = "isphone"
	PARAM_KW             = "kw"
	PARAM_KZ             = "kz"
	PARAM_LAST           = "last"
	PARAM_LZ             = "lz"
	PARAM_MARK           = "mark"
	PARAM_NET_TYPE       = "net_type"
	PARAM_OFFSET         = "offset"
	PARAM_PB_RN          = 20
	PARAM_PBRN           = "pb_rn"
	PARAM_PHONE_IMEI     = "_phone_imei"
	PARAM_PIC            = "pic"
	PARAM_PID            = "pid"
	PARAM_PN             = "pn"
	PARAM_QUOTE_ID       = "quote_id"
	PARAM_R              = "r"
	PARAM_RN             = "rn"
	PARAM_SIGN           = "sign"
	PARAM_SPID           = "spid"
	PARAM_SUG            = "q"
	PARAM_TBS            = "tbs"
	PARAM_TID            = "tid"
	PARAM_TITLE          = "title"
	PARAM_USER_PASSWORD  = "passwd"
	PARAM_USER_PHONE_NUM = "phonenum"
	PARAM_USER_SEX       = "sex"
	PARAM_USER_SMS_CODE  = "smscode"
	PARAM_USER_USERID    = "uid"
	PARAM_USER_USERNAME  = "un"
	PARAM_USER_VCODE     = "vcode"
	PARAM_USER_VCODE_MD5 = "vcode_md5"
	PARAM_WITHFLOOR      = "with_floor"
	POST_ADD             = 7
	PROFILE_MODIFY       = 0x10
	RmStore              = 0x23
	sCharAnd             = '&'
	sCharEqual           = '='
	sCharQuestion        = '?'
	SIGN_KEY             = "tiebaclient!!!"
	SYSTEM_FILLUNAME     = 0x1b
	SYSTEM_GET_SMS       = 0x1d
	SYSTEM_LOGIN         = 0x1a
	SYSTEM_Recommend     = 0x1f
	SYSTEM_REG           = 0x19
	SYSTEM_REG_REAL      = 30
	SYSTEM_SYNC          = 0x1c
	THREAD_ADD           = 6
	THREAD_COMMENT       = 11
	ThreadStore          = 0x24
	USER_FANS_PAGE       = 0x17
	USER_FEED_ATME       = 0x15
	USER_FEED_REPLYME    = 20
	USER_FOLLOW          = 13
	USER_FOLLOW_LIST     = 0x13
	USER_FOLLOW_PAGE     = 0x18
	USER_FOLLOW_SUG      = 0x12
	USER_PROFILE         = 0x16
	USER_UNFOLLOW        = 14
)

var BASE_URLS []string = []string{
	(HOST_ADDRESS + "f/forum/favolike"),   //0
	(HOST_ADDRESS + "f/frs/page"),         //1
	(HOST_ADDRESS + "f/pb/page"),          //2
	(HOST_ADDRESS + "f/pb/floor"),         //3
	(HOST_ADDRESS + "f/anti/vcode"),       //4
	(HOST_ADDRESS + "f/forum/sug"),        //5
	(HOST_ADDRESS + "c/thread/add"),       //6
	(HOST_ADDRESS + "c/post/add"),         //7
	(HOST_ADDRESS + "c/img/upload"),       //8
	(HOST_ADDRESS + "c/img/chunkupload"),  //9
	(HOST_ADDRESS + "c/img/finupload"),    //10
	(HOST_ADDRESS + "c/thread/comment"),   //11
	(HOST_ADDRESS + "c/forum/like"),       //12
	(HOST_ADDRESS + "c/user/follow"),      //13
	(HOST_ADDRESS + "c/user/unfollow"),    //14
	(HOST_ADDRESS + "c/forum/unlike"),     //15
	(HOST_ADDRESS + "c/profile/modify"),   //16
	(HOST_ADDRESS + "c/forum/unfavolike"), //17
	(HOST_ADDRESS + "u/follow/sug"),       //18
	(HOST_ADDRESS + "u/follow/list"),      //19
	(HOST_ADDRESS + "u/feed/replyme"),     //20
	(HOST_ADDRESS + "u/feed/atme"),        //21
	(HOST_ADDRESS + "u/user/profile"),     //22
	(HOST_ADDRESS + "u/fans/page"),        //23
	(HOST_ADDRESS + "u/follow/page"),      //24
	(HOST_ADDRESS + "s/reg"),              //25
	(HOST_ADDRESS + "s/login"),            //26
	(HOST_ADDRESS + "s/filluname"),        //27
	(HOST_ADDRESS + "s/sync"),             //28
	(HOST_ADDRESS + "s/getsmscode"),       //29
	(HOST_ADDRESS + "s/regreal"),          //30
	(HOST_ADDRESS + "s/recommendWin8"),    //31
	(HOST_ADDRESS + "c/forum/sign"),       //32
	(HOST_ADDRESS + "s/msg"),              //33
	(HOST_ADDRESS + "c/post/addstore"),    //34
	(HOST_ADDRESS + "c/post/rmstore"),     //35
	(HOST_ADDRESS + "f/post/threadstore"), //36
	(HOST_ADDRESS + "u/feed/mypost"),      //37
	(HOST_ADDRESS + "u/feed/otherpost"),   //38
	(HOST_ADDRESS + "c/img/portrait"),     //39
}

var 重试次数 int
var 最长允许响应时间 time.Duration

func INIT(_重试次数 int, _最长允许响应时间 time.Duration) {
	重试次数 = _重试次数

	最长允许响应时间 = _最长允许响应时间
}
