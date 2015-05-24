package kw_manager

import (
	//"sync"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/purstal/pbtools/modules/logs"

	"github.com/purstal/pbtools/tools-core/utils/action"
)

const (
	ChangeInterval action.Pattern = iota
	ChangeFile
)

type KeywordManager struct {
	Logger *logs.Logger

	FileName      string
	KewWordExps   []*regexp.Regexp
	LastModTime   time.Time
	CheckInterval time.Duration

	actChan chan action.Action
}

func NewKeywordManager(logger *logs.Logger) *KeywordManager {
	return &KeywordManager{Logger: logger, actChan: make(chan action.Action)}
}

func NewKeywordManagerBidingWithFile(keyWordFileFlieName string,
	checkInterval time.Duration, logger *logs.Logger) (*KeywordManager, error) {
	var m KeywordManager
	m.FileName = keyWordFileFlieName

	file, err := os.Open(m.FileName)
	if os.IsNotExist(err) {
		var err error
		file, err = os.Create(m.FileName)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else {
		err := LoadExps(file, &m.KewWordExps, logger)
		if err != nil {
			return nil, err
		}
	}

	if fi, err := file.Stat(); err != nil {
		return nil, err
	} else {
		m.LastModTime = fi.ModTime()
	}
	file.Close()

	m.CheckInterval = checkInterval
	m.actChan = make(chan action.Action)

	go func() {
		ticker := time.NewTicker(m.CheckInterval)
		for {
			select {
			case <-ticker.C:
			case act := <-m.actChan:
				switch act.Pattern {
				case ChangeInterval:
					ticker.Stop()
					ticker = time.NewTicker(act.Param.(time.Duration))
				case ChangeFile:
				}
				continue
			}
			file, err1 := os.Open(m.FileName)
			if err1 != nil {
				logger.Error("无法打开关键词文件,跳过本次.", err1, ".")
				continue
			}
			func() {
				defer func() { file.Close() }()
				fi, err2 := file.Stat()
				if err2 != nil {
					logger.Error("无法获取文件信息,跳过本次.", err2, ".")
					return
				}
				if modTime := fi.ModTime(); modTime != m.LastModTime {
					err := LoadExps(file, &m.KewWordExps, logger)
					if err != nil {
						logger.Error("无法更新关键词,下次修改前将不尝试读取.", err, ".")
					}
					m.LastModTime = modTime
				}
			}()
		}
	}()

	return &m, nil
}

func (m KeywordManager) ChangeCheckInterval(newInterval time.Duration) {
	m.actChan <- action.Action{ChangeInterval, newInterval}
}

func (m KeywordManager) ChangeKeyWordFile(newFile string) {
	m.actChan <- action.Action{ChangeFile, newFile}
}

func (m KeywordManager) KeyWords() []*regexp.Regexp {
	return m.KewWordExps
}

func LoadExps(file *os.File, exps *[]*regexp.Regexp, logger *logs.Logger) error {
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	lines := strings.Split(string(bytes), "\n")

	oldExps := make(map[string]*regexp.Regexp)

	for _, exp := range *exps {
		oldExps[exp.String()] = exp
	}

	newExps := make(map[string]*regexp.Regexp)
	var addedExps []string

	for lineNo, line := range lines {
		line = strings.TrimRightFunc(line, func(r rune) bool {
			return r == '\n' || r == '\r'
		})
		if line == "" {
			continue
		}
		if exp, exist := oldExps[line]; exist {
			newExps[line] = exp
			delete(oldExps, line)
		} else {
			newExp, err := regexp.Compile(line)
			if err != nil {
				logs.Error(fmt.Sprintf("不正确的关键词(第%d行),跳过.", lineNo), err)
			} else {
				newExps[line] = newExp
				addedExps = append(addedExps, line)
			}

		}
	}

	newExpSlice := make([]*regexp.Regexp, 0, len(newExps))

	for _, exp := range newExps {
		newExpSlice = append(newExpSlice, exp)
	}

	*exps = newExpSlice

	var updateInfo string = fmt.Sprintf("更新关键词(%s):\n", file.Name())
	for _, exp := range addedExps {
		updateInfo = updateInfo + "[+] " + exp + "\n"
	}
	for _, exp := range oldExps {
		updateInfo = updateInfo + "[-] " + exp.String() + "\n"
	}
	updateInfo = strings.TrimSuffix(updateInfo, "\n")
	logger.Info(updateInfo)

	//logger.Debug("现在的关键词:", newExpSlice, ".")

	return nil
}
