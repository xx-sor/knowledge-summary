## 1.countDocuments与estimatedDocumentCount
countDocuments是查询该集合下准确的文档总数；estimatedDocumentCount是查该集合下估算的文档总数。
### countDocuments 
countDocuments是扫描整个索引(如果过滤条件有索引的话)或者扫描整个集合，判断出符合条件的文档，将数量求和得到的。

### estimatedDocumentCount
estimatedDocumentCount是根据元数据估算来得到文档总数的。

### 性能差别
https://blog.csdn.net/fenglllle/article/details/120640744
600w条文档的时候，mongodb的2个查集合下总量的接口就已经是10倍的性能差距了：
237毫秒和3364毫秒
