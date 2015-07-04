package post_deleter

import (
	"os"
	"time"

	"github.com/purstal/pbtools/modules/logs"
	"github.com/purstal/pbtools/modules/postbar"

	postfinder "github.com/purstal/pbtools/tool-cores/post-finder"

	"github.com/purstal/pbtools/tools/simple-post-deleter/post-deleter/keyword-manager"
)

type PostDeleter struct {
	AccWin8, AccAndr *postbar.Account
	PostFinder       *postfinder.PostFinder

	Content_RxKw       *kw_manager.RegexpKeywordManager
	UserName_RxKw      *kw_manager.RegexpKeywordManager
	Tid_Whitelist      *kw_manager.Uint64KeywordManager
	UserName_Whitelist *kw_manager.StringKeywordManager
	BawuList           *kw_manager.StringKeywordManager

	ForumName string
	ForumID   uint64

	Records Records

	Logger   *logs.Logger
	OpLogger *OperationLogger
}

func NewPostDeleter(accWin8, accAndr *postbar.Account, forumName string, forumID uint64,
	content_RxKw_FileName, UserName_RxKw_FileName, Tid_Whitelist_FileName,
	UserName_Whitelist_FileName, BawuList_FileName string,
	debug bool, logDir string) *PostDeleter {
	var deleter PostDeleter
	var err error

	if !initLogger(&deleter, logDir) {
		return nil
	}

	deleter.OpLogger = NewOperationLogger(logDir)
	if deleter.OpLogger == nil {
		return nil
	}

	if deleter.Content_RxKw = newRxKwManager(content_RxKw_FileName, deleter.Logger); deleter.Content_RxKw == nil {
		return nil
	} else if deleter.UserName_RxKw = newRxKwManager(UserName_RxKw_FileName, deleter.Logger); deleter.UserName_RxKw == nil {
		return nil
	} else if deleter.Tid_Whitelist = newU64KwManager(Tid_Whitelist_FileName, deleter.Logger); deleter.Tid_Whitelist == nil {
		return nil
	} else if deleter.UserName_Whitelist = newStrKwManager(UserName_Whitelist_FileName, deleter.Logger); deleter.Tid_Whitelist == nil {
		return nil
	} else if deleter.BawuList = newStrKwManager(BawuList_FileName, deleter.Logger); deleter.Tid_Whitelist == nil {
		return nil
	}

	deleter.AccWin8, deleter.AccAndr = accWin8, accAndr
	deleter.ForumID, deleter.ForumName = forumID, forumName

	if deleter.PostFinder, err = postfinder.NewPostFinder(
		deleter.AccWin8, deleter.AccAndr, forumName,
		func(postfinder *postfinder.PostFinder) {
			postfinder.ThreadFilter = deleter.ThreadFilter
			postfinder.NewThreadFirstAssessor = deleter.NewThreadFirstAssessor
			postfinder.NewThreadSecondAssessor = deleter.NewThreadSecondAssessor
			postfinder.AdvSearchAssessor = deleter.AdvSearchAssessor
			postfinder.PostAssessor = deleter.PostAssessor
			postfinder.CommentAssessor = deleter.CommentAssessor
		}, debug, logDir); err != nil {
		return nil
	}

	deleter.Records.RulesThread_Tids, deleter.Records.ServerListThread_Tids,
		deleter.Records.WaterThread_Tids =
		map[uint64]struct{}{}, map[uint64]struct{}{}, map[uint64]struct{}{}

	return &deleter
}

func (p *PostDeleter) Run(monitorInterval time.Duration) {
	p.PostFinder.Run(monitorInterval)
}

func initLogger(pd *PostDeleter, logDir string) bool {
	logFile, err := os.Create(logDir + "post-deleter-日志.log")
	if err != nil {
		logs.Fatal("无法创建log文件.", err)
		return false
	}
	pd.Logger = logs.NewLogger(logs.DebugLevel, os.Stdout, logFile)
	logs.DefaultLogger = pd.Logger
	return true
}

func newRxKwManager(fileName string, logger *logs.Logger) *kw_manager.RegexpKeywordManager {
	var m *kw_manager.RegexpKeywordManager
	var err error
	if fileName != "" {
		m, err =
			kw_manager.NewRegexpKeywordManagerBidingWithFile(
				fileName, time.Second, logger)
		if err != nil {
			logger.Error("无法创建贴子内容正则Manager.", err)
			return nil
		}
		return m
	} else {
		logger.Warn("未设置正则关键词文件")
		return kw_manager.NewRegexpKeywordManager(logger)
	}
}

func newU64KwManager(fileName string, logger *logs.Logger) *kw_manager.Uint64KeywordManager {
	var m *kw_manager.Uint64KeywordManager
	var err error
	if fileName != "" {
		m, err =
			kw_manager.NewUint64KeywordManagerBidingWithFile(
				fileName, time.Second, logger)
		if err != nil {
			logger.Error("无法创建贴子内容正则Manager.", err)
			return nil
		}
		return m
	} else {
		logger.Warn("未设置正则关键词文件")
		return kw_manager.NewUint64KeywordManager(logger)
	}
}

func newStrKwManager(fileName string, logger *logs.Logger) *kw_manager.StringKeywordManager {
	var m *kw_manager.StringKeywordManager
	var err error
	if fileName != "" {
		m, err =
			kw_manager.NewStringKeywordManagerBidingWithFile(
				fileName, time.Second, logger)
		if err != nil {
			logger.Error("无法创建贴子内容正则Manager.", err)
			return nil
		}
		return m
	} else {
		logger.Warn("未设置正则关键词文件")
		return kw_manager.NewStringKeywordManager(logger)
	}
}
