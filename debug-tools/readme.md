一个简单的客户端api测试软件

运行:

	debug-tools $port
其中 $port  是端口号
默认端口是33120

打开浏览器,可以通过`localhost:$port`直接测试客户端api

##usage/例子:##
(先通过`debug-tools 33120`或直接`debug-tools`运行)

	http://localhost:33120/c/f/forum/favolike?client=Win8&fmt_json&BDUSS=xxx
* client:
	* andr: 安卓版`6.1.2`
	* win8: windows8版`1.5.0.0`
	* costum: 自定义
		* `net_type`
		* `_client_type`
		* `_client_id`
		* `_client_version`
* fmt_json: 格式化响应的json
* BDUSS: BDUSS
* 其他参数跟在地址后就行