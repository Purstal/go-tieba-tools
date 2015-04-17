package thread

import (
	//"fmt"
	//"os"
	"strconv"
	//"time"

	"github.com/purstal/pbtools/modules/logs"
	"github.com/purstal/pbtools/modules/postbar"
)

func ExtractContent(originalContentList []interface{}) []postbar.Content {

	var contents []postbar.Content = make([]postbar.Content, 0)

	for _, originalContent := range originalContentList {
		var contentMap map[string]interface{}
		var ok bool
		if contentMap, ok = originalContent.(map[string]interface{}); !ok {
			logs.Error("获取内容中的一项失败", originalContent)
			contents = append(contents, originalContent)
			continue
		}

		var content postbar.Content

		func() {
			defer func() {
				err := recover()
				if err != nil {
					logs.Error("获取内容中的一项的属性失败:", err, contentMap)
					content = contentMap
				}
			}()

			var contentType int32
			switch contentMap["type"].(type) {
			case (string):
				contentType_str, _ := strconv.Atoi(contentMap["type"].(string))
				contentType = int32(contentType_str)
			case (float64):
				contentType = int32((contentMap["type"].(float64)))
			}
			switch contentType {
			case 0: //文字,楼中楼会把所在楼层的图片也归为文字.
				content = postbar.Text{contentMap["text"].(string)}
			case 1: //链接
				content = postbar.Link{contentMap["link"].(string), contentMap["text"].(string)}
			case 2: //表情
				content = postbar.Emoticon{contentMap["text"].(string), contentMap["c"].(string)}
			case 3: //图片
				content = postbar.Pic{contentMap["src"].(string)}
			case 4: //@
				var uidStr = contentMap["uid"].(string)
				var uid, _ = strconv.ParseUint(uidStr, 10, 64)
				content = postbar.At{contentMap["text"].(string), uid}
			case 5: //音乐??视频??
				content = postbar.Video{contentMap["text"].(string)}
			case 6: //视频??
			case 10: //语音
				during_time, _ := strconv.Atoi(contentMap["during_time"].(string))
				content = postbar.Voice{int32(during_time), contentMap["voice_md5"].(string)}
			case 11: //表情商店里的表情 //没什么好搞的
				content = contentMap
			default:
				content = contentMap
			}
		}()

		contents = append(contents, content)

	}
	return contents

}
