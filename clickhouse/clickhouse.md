# 本地表与分布式表
在 ClickHouse 集群里，真正存数据的是每台机器上的“本地表”（通常是 ReplicatedMergeTree 家族）。
“分布式表”只是一个路由/聚合层，本身不存数据，只负责把查询或写入按照集群拓扑分发到各个分片上的本地表，再汇总结果。类似于mysql分片的proxy。
## 架构理念

## 分区与分片
### 分片
由clusters.xml定义集群层：对应集群有几个分片，每个分片对应的机器分别是哪个。
由分布式表的建表语句定义分片key，即定义路由规则。

### 分区
分区是ClickHouse 本地表（如 MergeTree）内部的数据组织方式，按照某个字段（或表达式）对数据进行逻辑分组存储。
目的：
(1)加快查询效率（只查你需要的分区）
(2)加快删除效率（DROP PARTITION）
(3)控制数据生命周期（TTL 按分区清理）


# 建表
## 常见做法
1. 配置集群拓扑
在 config.xml 或 clusters.xml 中定义 <remote_servers>，告诉 ClickHouse：
有哪些集群（cluster）
每个集群有几个分片（）
每个分片有几个副本（）
每个副本的机器地址、端口等信息
示例配置：
```xml
<remote_servers>
    <my_cluster>
        <shard>
            <replica>
                <host>node1</host>
                <port>9000</port>
            </replica>
            <replica>
                <host>node2</host>
                <port>9000</port>
            </replica>
        </shard>
        <shard>
            <replica>
                <host>node3</host>
                <port>9000</port>
            </replica>
            <replica>
                <host>node4</host>
                <port>9000</port>
            </replica>
        </shard>
    </my_cluster>
</remote_servers>
```
注意：
这里的分片指的是集群层的分片，并无逻辑意义，逻辑意义由分布式表的建表语句的sharding key赋予。
即：这里只指定有几个分片，每个分片分别在哪台机器上。不指定路由规则和数据的分片规则。

### 集群信息查看
```sql 
SELECT * FROM system.clusters\G;

Row 1:
──────
cluster:                 default_cluster
shard_num:               1
shard_weight:            1
replica_num:             1
host_name:               x.x.xx.x
host_address:            x.x.xx.x
port:                    9000
is_local:                1
user:                    default
default_database:        
errors_count:            0
slowdowns_count:         0
estimated_recovery_time: 0
database_shard_name:     
database_replica_name:   
is_active:               ᴺᵁᴸᴸ

Row 2:
──────
cluster:                 default_cluster
shard_num:               2
shard_weight:            1
replica_num:             1
host_name:               x.x.xx.xx
host_address:            x.x.xx.xx
port:                    9000
is_local:                0
user:                    default
default_database:        
errors_count:            0
slowdowns_count:         0
estimated_recovery_time: 0
database_shard_name:     
database_replica_name:   
is_active:               ᴺᵁᴸᴸ

2 rows in set. Elapsed: 0.002 sec. 

```


2. 建本地表：
在每个分片上建 ReplicatedMergeTree 本地表（用宏 {shard}/{replica}）。可以用 ON CLUSTER 一条语句在所有节点创建相同的库表结构。（本质上仍是在各节点创建本地表）
示例:
```sql 
CREATE TABLE my_local_table ON CLUSTER my_cluster
(
    id UInt64,
    name String,
    ...
)
ENGINE = ReplicatedMergeTree(
    '/clickhouse/tables/{shard}/my_local_table',  -- ZooKeeper 路径模板
    '{replica}'                                    -- 每个副本的唯一标识（由 ClickHouse 自动替换）
)
ORDER BY id;
```
说明：
(1) {shard}：会自动替换为该节点所属的分片编号（shard1、shard2…）。()(2) {replica}：会自动替换为该副本的唯一标识（通常是主机名）。
(3) ClickHouse会根据集群配置，自动把这条 SQL 分发到所有机器上，创建对应的本地表结构。

3. 建分布式表:
建一个 Distributed 表指向上面的本地表，指定 cluster 和分片键。
示例：
```sql
CREATE TABLE my_distributed_table ON CLUSTER my_cluster
AS my_local_table
ENGINE = Distributed(my_cluster, default, my_local_table, cif);
```

如果只建了分布式表而没有对应的目标本地表，读写都会失败，因为没有实际存储可用。

## 不像集群版mysql那样隐藏"本地表"这个概念的原因
1. 存储布局对 OLAP 性能至关重要：分区键、ORDER BY（数据跳过索引）、TTL、存储策略、压缩、索引颗粒度等，都是在本地 MergeTree 上定义并发挥作用的。让代理“自动决定”往往会错，或者丧失可控性。
2. 节点特异参数无法被一个“全局表定义”统一描述：复制路径（Keeper 路径）、replica 名称、磁盘/卷策略、冷热分层等每个副本都不同，需要在各节点落地时用 {shard}/{replica} 宏展开，这天然就是“本地对象”。
3. 一对多的路由视图：同一批本地表之上，可能需要建立多张分布式表（不同的分片键、不同的只读/只写路由、不同的采样策略）以适配多种查询/写入模式。这要求存储层与路由层解耦。
4. 架构哲学差异：MySQL 常见是“复制优先、分片靠中间件隐式屏蔽”的 OLTP 思路；ClickHouse 是“共享无 + 显式分布”的 OLAP 思路，鼓励正视并利用数据分布来拿性能。


# 删表

