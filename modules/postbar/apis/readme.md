#api列表

##目录

* [account.go](#accountgo)
	* [RLogin](#rlogin)(acc, userName, password)
		* [Login](#login)(acc, password)
	* [IsLogin](#islogin)(BDUSS)
* [bawu.go](#bawugo)
	* [DeletePost](#deletepost)(acc, pid)
	* [DeleteThread](#deletethread)(acc, tid)
	* [BlockIDWeb](#blockidweb)(BDUSS, forumID, userName, pid, day, reason)
	* [CommitPrison](#commitprison)(acc, forumName, forumID, userName, threadID, postID, day ,reason)
* [message.go](#messagego)
	* [RFeedReplyMe](#rfeedreplyme)(acc)
		* https://github.com/Purstal/pbtools/tree/master/modules/postbar/apis/message/
	* [RFeedAtMe](#rfeedatme)(acc)
	* [GetNotice](#getnotice)(acc)
* [misc.go](#miscgo)
	* [RSearchForum](#rsearchforum)(forumName)
		* [SearchForum](#searchforum)(forumName)
* [page.go](#pagego)
	* [RGetForum](#rgetforum)(acc, kw, rn, pn)
		* https://github.com/Purstal/pbtools/tree/master/modules/postbar/apis/forum-win8-1.5.0.0/
	* [RGetThread](#rgetthread)(acc, tid, mark, pid, pn, rn, withFloor, seeLz, r)
		* https://github.com/Purstal/pbtools/tree/master/modules/postbar/apis/thread-win8-1.5.0.0/
	* [RGetFloor](#rgetfloor)(acc, tid, isComment, id, pn)
		* https://github.com/Purstal/pbtools/tree/master/modules/postbar/apis/floor-andr-6.1.3/
* [post.go](#postgo)
	* [AddPost](#addpost)(accAndr, content, fid, forumName, tid, floorNumber, quoteID)
* [sign-in.go](#sign-ingo)
	* [GetForumList](#getforumlist)(accAndr)
* [special.go](#specialgo)
	* [GetFid](#getfid)(forumName)
	* [GetUid](#getuid)(userName)
	* [RGetTbs](#rgettbs)(acc)
		* [GetTbs](#gettbs)(acc)
	* [RGetTbsWeb](#rgettbsweb)(BDUSS)
		* [GetTbsWeb](#gettbsweb)(BDUSS)
* [unknown.go](#unknowngo)
	* [RTest](#rtest)(acc, userName, password)
* [users.go](#usersgo)
	* [GetUserForumLike](#getuserforumlike)(acc, uid)
	* [GetUserInfo](#getuserinfo)(acc, uid)

---
* 以R为前缀的函数返回 `([]byte, error)`
* 非以R为前缀的函数返回封装过的相关的东西,一般都还包含`error`和`*postbar.PbError`

##accounts.go
###RLogin
	func RLogin(acc *postbar.Account, userName, password string) ([]byte, error)
####Login
	func Login(acc *postbar.Account, password string) (error, *postbar.PbError)
验证码请自行解决╮(╯▽╰)╭~

###IsLogin
	func IsLogin(BDUSS string) (bool, error)
很明显,用来检验是否登录.
其实就是取网页端tbs,响应里有个`is_login`可以检验是否登录.

##bawu.go
###DeletePost
	func DeletePost(acc *postbar.Account, pid uint64) (error, *postbar.PbError)
吧务删贴,主题也可以(用一楼的pid).

###DeleteThread
	func DeleteThread(acc *postbar.Account, tid uint64) (error, *postbar.PbError)
吧务删主题.

###BlockIDWeb
	func BlockIDWeb(BDUSS string, forumID uint64, userName string, pid uint64, day int, reason string) (error, *postbar.PbError)
网页版的封禁.

###CommitPrison
	func CommitPrison(account *postbar.Account, forumName string, forumID uint64, userName string, threadID, postID uint64, day int, reason string) (error, *postbar.PbError)
客户端的封禁,曾经百度加了参数使之失效了,后来我补上了,但还没有测试过.要用推荐用上面那个.

##message.go
###RFeedReplyMe
	func RFeedReplyMe(acc *postbar.Account) ([]byte, error)
回复的消息提醒.
`./message/`有个封装.

###RFeedAtMe
	func RFeedAtMe(acc *postbar.Account) ([]byte, error)
@的消息提醒.

###GetNotice
	func GetNotice(acc *postbar.Account) (*Notice, error, *postbar.PbError)
获取各类提醒的数量.

##misc.go
###RSearchForum
	func RSearchForum(forumName string) ([]byte, error)
####SearchForum
	func SearchForum(forumName string) ([]ForumSearchResult, error, *postbar.PbError)
搜索贴吧,主要目的是用来获取fid.一开始我以为`GetFid() #special.go`失效了才打算使用这个作为替代品,后来发现原来没有实效...


##page.go
###RGetForum
	func RGetForum(acc *postbar.Account, kw string, rn,pn int) ([]byte, error)
获取主页,封装在`./forum-win8-1.5.0.0/`.

###RGetThread
	func RGetThread(acc *postbar.Account, tid uint64, mark bool, pid uint64, pn, rn int, withFloor, seeLz, r bool) ([]byte, error)
获取主题页,封装在`./thread-win8-1.5.0.0/`.

###RGetFloor
	func RGetFloor(acc *postbar.Account, tid uint64, isComment bool, id uint64, pn int) ([]byte, error)
获取楼中楼页,封装在`./floor-andr-6.1.3/`.

##post.go
###AddPost
	func AddPost(accAndr *postbar.Account, content string, fid uint64, forumName string, tid uint64, floorNumber int, quoteID uint64) (error, *postbar.PbError)
回贴.

##sign-in.go
###GetForumList
	func GetForumList(accAndr *postbar.Account) (*ForumList, error, *postbar.PbError)
获取一键签到的贴吧列表,非会员最多50个,会员最多200个.

##special.go
###GetFid
	func GetFid(forumName string) (uint64, error, *postbar.PbError)
通过贴吧名获取贴吧fid.

###GetUid
	func GetUid(userName string) (uint64, error)
通过用户名获取用户uid,非客户端api,少数不是自己亲自找到的api.
有的用户用这个取不到uid,试试搜用户的api吧(>_>我没弄).

###RGetTbs
	func RGetTbs(acc *postbar.Account)
####GetTbs
	func GetTbs(acc *postbar.Account) (string, error, *postbar.PbError)
获取tbs,客户端api,蒙出来的api,至少安卓客户端和win8客户端都没用上这个.

###RGetTbsWeb
	func RGetTbsWeb(BDUSS string) ([]byte, error)
####GetTbsWeb
	func GetTbsWeb(BDUSS string) (string, error)
获取tbs,网页端的api,少数不是自己亲自找到的api.

##unknown.go
###RTest
	func RTest(acc *postbar.Account, userName, password string) ([]byte, error)
蒙出来的api...
╮(╯▽╰)╭鬼知道什么玩意,但是参数和登陆一样.

##users.go
###GetUserForumLike
	func GetUserForumLike(acc *postbar.Account, uid uint64) (*ForumLikePageForumList, error)
获取用户关注的贴吧,少数不是自己亲自找到的api.

###GetUserInfo
	func GetUserInfo(acc *postbar.Account, uid uint64) (*UserInfo, error)
干嘛用的来的..貌似只是用来从用户uid获得用户名的...

##其他
	以前还有点赞,发主题,置顶/取消置顶,加精/取消加精等api,后来觉得没什么用就扔了...
	不过确实意义不大,除了点赞外其他都要跟验证码甚至手机验证打交道吧...