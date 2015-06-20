package inireader

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

import "github.com/purstal/pbtools/tool-core/operation-analyser/old/log"

func ReadINI(name string) (INI, error) {
	var f *os.File
	var err error
	if f, err = os.Open(name); err != nil {
		return nil, err
	}

	all_bs, _ := ioutil.ReadAll(f)
	all := string(all_bs)
	all = strings.TrimLeft(all, string([]byte{0xEF, 0xBB, 0xBF}))
	all = strings.Replace(all, "\r", "", -1)
	lines := strings.Split(all, "\n")

	ini := make(map[string]map[string]string)
	var where string = ""

	for i, line := range lines {

		if len(line) == 0 {
			continue
		}
		switch {

		case strings.HasPrefix(line, "["):
			{
				where = line[1:strings.Index(line, "]")]
				ini[where] = make(map[string]string)
			}
		case strings.HasPrefix(line, ";"):
			{
				continue
			}
		default:
			{
				slice := strings.Split(line, "=")
				if len(slice) != 2 {
					log.Loglog("ini语法错误于第" + strconv.Itoa(i) + "行:[ " + line + " ],也许是忘了写注释符号?此行已被跳过")
					continue
				}
				slice2 := strings.Split(slice[1], ";")
				ini[where][slice[0]] = slice2[0]
			}

		}
	}
	return ini, nil

}

type INI map[string]map[string]string
