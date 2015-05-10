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