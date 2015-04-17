package misc

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"os"
	"strconv"
	"time"

	"code.google.com/p/mahonia"

	"github.com/purstal/pbtools/modules/logs"
)

func ComputeBase64(str string) (res string) {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func ComputeMD5(src string) (res string) {
	h := md5.New()
	h.Write([]byte(src))
	return hex.EncodeToString(h.Sum(nil))
}

func WriteNewFile(dir, fn, c string) {
	os.MkdirAll(dir, 0777)
	f, err := os.Create(dir + fn)
	if err != nil {
		logs.Error("创造文件时失败", err.Error())
	} else {
		f.WriteString(c)
	}
	if f != nil {
		f.WriteString(c)
	}

}

func ToGBK(src string) (dst string) {
	//encoder := mahonia.NewEncoder("gbk")
	encoder := mahonia.NewEncoder("gb2312")
	return encoder.ConvertString(src)
}

func StringSliceToGBK(src []string) (dst []string) {
	dst = make([]string, len(src))
	for i, str := range src {
		dst[i] = ToGBK(str)
	}
	return dst
}

func StringSliceFromGBK(src []string) (dst []string) {
	dst = make([]string, len(src))
	for i, str := range src {
		dst[i] = FromGBK(str)
	}
	return dst
}

func FromGBK(src string) (dst string) {
	decoder := mahonia.NewDecoder("gbk")
	return decoder.ConvertString(src)
}

func UrlQueryEscape(s string) string {
	spaceCount, hexCount := 0, 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if shouldEscape(c) {
			if c == ' ' {
				spaceCount++
			} else {
				hexCount++
			}
		}
	}

	if spaceCount == 0 && hexCount == 0 {
		return s
	}

	t := make([]byte, len(s)+2*hexCount)
	j := 0
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == ' ':
			t[j] = '+'
			j++
		case shouldEscape(c):
			t[j] = '%'
			t[j+1] = "0123456789ABCDEF"[c>>4]
			t[j+2] = "0123456789ABCDEF"[c&15]
			j += 3
		default:
			t[j] = s[i]
			j++
		}
	}
	return string(t)
}

func shouldEscape(c byte) bool {
	// §2.3 Unreserved characters (alphanum)
	if 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || '0' <= c && c <= '9' {
		return false
	}

	if c == '-' || c == '_' || c == '.' || c == '~' {
		return false
	}

	// Everything else must be escaped.
	return true
}

func TryCreateFile(name string) *os.File {
	var err1, err2 error
	var f *os.File
	f, err1 = os.Create(name + ".")
	if err1 != nil {
		logs.Warn(name+".",
			"创建失败:\n", err1.Error(), "\n尝试创建为",
			name+strconv.FormatInt(time.Now().Unix(), 10)+".")
		f, err2 = os.Create(name + strconv.FormatInt(time.Now().Unix(), 10) + ".")
		if err2 != nil {
			logs.Fatal(name+".",
				"创建失败:\n", err1.Error(), "\n跳过写入此文件")
		}
	}
	return f
}
