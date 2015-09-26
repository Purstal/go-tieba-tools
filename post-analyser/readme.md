#usage#
##post-scaner##
    post-scaner $tid $from-time $to-time

* $from-time $to-time: 如 2006-01-02.
* 生成的json文件在`scanned-thread`目录.
* json文件没压缩,会很大.

##post-analyser##
    post-analyser $file
* $file需用unix风格.
* 生成的csv不带bom头,直接用excel打开会乱码,打开前请自行添加bom头.
* 生成的文件在`analyse-result`目录.

##a##
    a $from-time $to-time
* 简单地扫描在thread-list里的每行第一个捕获到的`(\d+)`,当做tid执行`post-scaner $tid $from-time $to-time`.
* 如果没捕获到就跳过.

##b##
    b
* 没错,就这么简单...
* 把所有`scanned-thread`目录下的文件执行一遍`post-analyser $filename`.
* 如果没有错误,post-analyser是没有输出的...