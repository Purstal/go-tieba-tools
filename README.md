个人有关百度贴吧的一些工具,比如用来监控贴吧<br/>
因为算是自己第一个能拿的出手(?)的东西,所以肯定有许多不足,请见谅.<br/>
缓速搬运中...<br/>

#导航(?)#

##API##

###**处理客户端的验证:**###
https://github.com/Purstal/pbtools/blob/master/modules/postbar/signture.go

###**非浏览类的api:**###
https://github.com/Purstal/pbtools/tree/master/modules/postbar/apis

###**浏览类的api:**###

* 主页:
https://github.com/Purstal/pbtools/tree/master/modules/postbar/forum-win8-1.5.0.0/forum.go
* 主题:
https://github.com/Purstal/pbtools/tree/master/modules/postbar/thread-win8-1.5.0.0/forum.go
* 楼层:
https://github.com/Purstal/pbtools/tree/master/modules/postbar/floor-andr-6.1.3/floor.go

##工具##

###**简易的删贴机:**###
可以按照自己的需求修改
####核心部分:####
* 首页监控:
https://github.com/Purstal/pbtools/tree/master/tools-core/forum-page-monitor
* 贴子查找:
https://github.com/Purstal/pbtools/tree/master/tools-core/post-finder

####简单的本体:####
V ?.?
https://github.com/Purstal/pbtools/tree/master/tools/simple-post-deleter

###**操作量统计:**###
V 1.7
新版懒得写分析部分了,要让软件分析可以`-use-old-version`.
但是每次分析都要扫描一遍.<br/>
新版会自动把扫描过的保存在本地,但是是按年保存的,所以读取得耗点时间,自己写的时候经历过`直接保存(某天的json)`->`保存(某天的json).tar.gz,运行软件自动json->json.tar.gz`->`保存(每年的jsons).tar.gz,运行软件自动((某天的json)|(某天的json).tar.gz)->(每年的jsons).tar.gz`,懒得再弄成每月一`.tar.gz`了,就这样吧...<br/>
顺便,自己想分析可以改operation-analyser.go里的`analyse`函数,甚至可以在那里把格式转成旧的格式(其实基本上没什么变话),然后调用`ProcDelete`以及`WriteToCSV_bawu`,如果看得懂怎么用的话..我猜的....记得先把各个吧务的删贴记录分离出来,再把全部的恢复记录分出来,那个`ProcDelete`要用到..

其实我是想把这个放到另一个repository的..但是懒得去了..因为其实接下来的打算是弄`operation-analyser analyse --pbfile=minecraft.toml lastyear range[2015-4-1,yesterday] 2015-2`类似的东西的,但因为这个项目优先级不高,就先(?)不填了..

####核心部分:####
https://github.com/Purstal/pbtools/tree/master/tools-core/operation-analyser
####本体:####
https://github.com/Purstal/pbtools/tree/master/tools/operation-analyser

#依赖#
github.com/PuerkitoBio/goquery<br/>
code.google.com/p/mahonia<br/>
github.com/BurntSushi/toml<br/>


<!--兔子我喜欢你!-->