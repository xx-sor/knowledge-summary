# 订阅模式
Exclusive	一个订阅只能有一个消费者
Failover	主备模式（按 partition 分配）
Shared	多个消费者并发消费同一个订阅（支持并发）
Key_Shared	按 key 保证顺序，支持并发消费

# 消费机制和提交顺序
## kafka
对同一个分区的消息，kafka客户端会根据offset的顺序拉取消息。
只有当offset较小的消息被拉取并处理后，才会拉取offset较大的消息并处理。（注意这里是不管是否提交的。）
当offset较大的消息被提交后，就认为小于这个offset的消息均已提交。

### 实现细节
每个 Consumer Group 对每个 Partition 只维护一个单一的「committed offset」。这个 offset 是由 Consumer 提交到 Kafka 的 __consumer_offsets 主题中。

### 为什么 Kafka 要这样设计？
1. 性能优先：
目标是高吞吐，低延迟。如果记录每条消息的ack状态，会降低性能。


2. 简化broker设计
kafka的broker更像是一个「分布式日志存储系统」(理念是顺序写，顺序读)，不做太多状态管理。
它只提供 offset 存储（提交）功能（供consumer读取、判断），不做消费状态追踪。


#### 如果kafka要记录每条消息的ack状态
1. 需要为每个 partition 增加一个 bitmap 或集合来维护「ack 状态」。
且Broker 需要频繁更新这些状态。
2. 放大了IO的写次数
每次 ack 一条消息，需要更新对应的 ack 状态（硬盘随机写，影响性能）。
3. 高并发下存在锁竞争：
多个 Consumer 并发 ack 消息，可能同时修改同一个分区的 ack 状态。
Broker 需要加锁或引入线程安全机制，来处理并发 ack 状态更新。
还要考虑消费组的 rebalance 时的状态一致性。
这些会大大增加 Broker 内部的状态管理复杂度。
(todo:那pulsar是怎么解决这个问题的？)

## pulsar
### pulsar是怎么解决每条消息都需要ACK而带来的额外问题的？
pulsar为了支持每条消息可以ack，给broker引入了状态。带来了以下问题：
1. 性能问题，每条消息都要确认，还存在并发时有锁竞争，导致性能进一步降低。
2. IO次数被放大
3. 消费者组rebalance或failover时的状态一致性

#### 性能问题解决
##### IO问题
1. ack操作放在内存中，ack请求不会立刻写磁盘，而是异步更新，后台线程会定期把 ack 状态同步到 BookKeeper（持久化）:
Pulsar 使用 Apache BookKeeper 存储消息日志（data）和 ledger metadata（ack、cursor 等）。
每个订阅（subscription）会维护一个 AckedMessageTracker 和一个 Cursor。
Cursor 的更新是顺序追加（append-only），而不是随机写。并且写入是异步、批量的（由 Broker 定时 flush）。

###### AckedMessageTracker

###### Cursor


##### 性能问题
2. 批量ACK合并(ack grouping)：
Pulsar 提供了 ack grouping 参数，允许客户端：
定时（如 100ms）或每 N 条消息后批量发送 ack
Broker 接收到的是一个“ack range”或“ack bitmap”而不是单条 ack。

##### 消费者组 rebalance / failover 时的状态一致性


# ledger


