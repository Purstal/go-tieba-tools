package caozuoliang

import (
	"code.google.com/p/mahonia"
	"os"
)

func WriteNewFile(fn, c string) {
	f, _ := os.Create(fn)
	defer f.Close()

	f.WriteString(c)
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
