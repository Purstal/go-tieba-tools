#不再更新了,明年(2016)就高考了w.
---

**突然发觉,部分依赖时间输入的程序在非东八时区的电脑使用会造成问题,而有时间输出的程序也会出现混乱.当然这些还没有测试..**

个人有关百度贴吧的一些工具,比如用来监控贴吧<br/>
因为算是自己第一个能拿的出手(?)的东西,所以肯定有许多不足,请见谅.<br/>


#导航(?)#

##API##

###**处理客户端的验证:**###
https://github.com/Purstal/pbtools/blob/master/modules/postbar/signture.go

###**apis:**###
https://github.com/Purstal/pbtools/tree/master/modules/postbar/apis


##工具##

###简易的删贴机:###
可以按照自己的需求修改
####核心部分:####
* 首页监控:
https://github.com/Purstal/pbtools/tree/master/tool-cores/forum-page-monitor
* 贴子查找:
https://github.com/Purstal/pbtools/tree/master/tool-cores/post-finder

####简单的本体:####
V ?.?
https://github.com/Purstal/pbtools/tree/master/tools/simple-post-deleter

###操作量统计:###
V 1.7
新版懒得写分析部分了,要让软件分析可以`-use-old-version`.
但是每次分析都要扫描一遍.<br/>
新版会自动把扫描过的保存在本地,按月保存的,读取耗点时间.<br/>
顺便,自己想分析可以改operation-analyser.go里的`analyse`函数,甚至可以在那里把格式转成旧的格式(其实基本上没什么变话),然后调用`ProcDelete`以及`WriteToCSV_bawu`,如果看得懂怎么用的话..我猜的....记得先把各个吧务的删贴记录分离出来,再把全部的恢复记录分出来,那个`ProcDelete`要用到..

本来我是想把这个放到另一个repository的,但是懒得去了..其实接下来的打算是弄`operation-analyser analyse --pbfile=minecraft.toml lastyear range[2015-4-1,yesterday] 2015-2`类似的东西的,但因为这个项目优先级不高,就先(?)不填了..

####核心部分:####
https://github.com/Purstal/pbtools/tree/master/tool-cores/operation-analyser
####本体:####
https://github.com/Purstal/pbtools/tree/master/tools/operation-analyser

#依赖#
github.com/PuerkitoBio/goquery<br/>
code.google.com/p/mahonia<br/>
github.com/BurntSushi/toml<br/>
github.com/shiena/ansicolor<br/>

###其他:###

见 https://github.com/Purstal/pbtools/tree/master/tools




<!--5YWU5a2Q5oiR5Zac5qyi5L2gIQ==-->