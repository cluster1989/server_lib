# conf

说明
=====

配置文件的解析类

先设定一个对应的类
然后写以下config文件
读取的时候，会直接对应到conf文件
-----

conf.Parse("filename")
解析文件

conf.Reload()
重新加载这个文件

conf.Unmarshal(obj)
将对象序列化到对应的配置类中

详细请参考test文件