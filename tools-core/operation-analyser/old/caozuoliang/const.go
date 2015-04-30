package caozuoliang

import (
	"time"
)

import "github.com/purstal/pbtools/tools-core/operation-analyser/old/net"

const Version = "1.6.2"

const (
	无    OPtype = "0"
	删贴   OPtype = "12"
	恢复   OPtype = "13"
	加精   OPtype = "17"
	取消加精 OPtype = "18"
	置顶   OPtype = "25"
	取消置顶 OPtype = "26"
)

const (
	主题 int = 0
	回复 int = 1
)

var 主题_GBK = ToGBK("主题")
var 回复_GBK = ToGBK("回复")

func StringPostType(posttype int) string {
	if posttype == 主题 {
		return 主题_GBK
	}
	return 回复_GBK
}

func EnsureYear(tid int) int {
	if tid > TheTidOfTheFirstPostOf2014 {
		return 2014
	} else if tid > TheTidOfTheFirstPostOf2013 {
		return 2013
	} else if tid > TheTidOfTheFirstPostOf2012 {
		return 2012
	}
	return 2011
}

func EnsureOperation(ostr string) OPtype {
	switch ostr {
	case "删贴":
		return 删贴
	case "恢复":
		return 恢复
	case "加精":
		return 加精
	case "取消加精":
		return 取消加精
	case "置顶":
		return 置顶
	case "取消置顶":
		return 取消置顶
	}
	return 无
}

func StringOpType(op_type OPtype) string {
	switch op_type {
	case 删贴:
		return "删贴"
	case 恢复:
		return "恢复"
	case 加精:
		return "加精"
	case 取消加精:
		return "取消加精"
	case 置顶:
		return "置顶"
	case 取消置顶:
		return "取消置顶"
	}
	return ""
}

//1m,2m,3m,4m,5m,6m,7m,8m,9m,10m,12m,15m,18m,20m,
//25m,30m,40m,50m,1h,1.5h,2h,2.5h,3h,4h,5h,6h,12h,
//1d,2d,3d,5d,10d,30d

var SpeedClass = []time.Duration{
	time.Minute * 1,
	time.Minute * 2,
	time.Minute * 3,
	time.Minute * 4,
	time.Minute * 5,
	time.Minute * 6,
	time.Minute * 7,
	time.Minute * 8,
	time.Minute * 9,
	time.Minute * 10,
	time.Minute * 12,
	time.Minute * 15,
	time.Minute * 18,
	time.Minute * 20,
	time.Minute * 25,
	time.Minute * 30,
	time.Minute * 40,
	time.Minute * 50,
	time.Hour * 1,
	time.Minute * 90,
	time.Hour * 2,
	time.Minute * 150,
	time.Hour * 3,
	time.Hour * 4,
	time.Hour * 5,
	time.Hour * 6,
	time.Hour * 7,
	time.Hour * 8,
	time.Hour * 9,
	time.Hour * 10,
	time.Hour * 12,
	time.Hour * 15,
	time.Hour * 18,
	time.Hour * 21,
	time.Hour * 24 * 1,
	time.Hour * 24 * 2,
	time.Hour * 24 * 3,
	time.Hour * 24 * 5,
	time.Hour * 24 * 8,
	time.Hour * 24 * 10,
	time.Hour * 24 * 15,
	time.Hour * 24 * 20,
	time.Hour * 24 * 25,
	time.Hour * 24 * 30,
}

var SpeedClassString = []string{
	"1分钟",
	"2分钟",
	"3分钟",
	"4分钟",
	"5分钟",
	"6分钟",
	"7分钟",
	"8分钟",
	"9分钟",
	"10分钟",
	"12分钟",
	"15分钟",
	"18分钟",
	"20分钟",
	"25分钟",
	"30分钟",
	"40分钟",
	"50分钟",
	"1小时",
	"1.5小时",
	"2小时",
	"2.5小时",
	"3小时",
	"4小时",
	"5小时",
	"6小时",
	"7小时",
	"8小时",
	"9小时",
	"10小时",
	"12小时",
	"15小时",
	"18小时",
	"21小时",
	"1天",
	"2天",
	"3天",
	"5天",
	"8天",
	"10天",
	"15天",
	"20天",
	"25天",
	"30天",
}

var 旧贴间隔 time.Duration
var 同账号判定 int
var 同主题判定 int
var retry_times int //3

func INIT(_重试次数 int, _最长允许响应时间 time.Duration, c, d, e int) {
	net.INIT(_重试次数, _最长允许响应时间)
	旧贴间隔 = time.Hour * time.Duration(c)
	同账号判定 = d
	同主题判定 = e

}

const (
	TheTidOfTheFirstPostOf2012 = 1347023954
	TheTidOfTheFirstPostOf2013 = 2076654371
	TheTidOfTheFirstPostOf2014 = 2790164985
)

//以2000年为起始
