# 订阅模式
1. Exclusive	一个订阅只能有一个消费者
2. Failover	主备模式（按 partition 分配）
每个 partition 同时只由一个 consumer 消费。多个消费者可以分别消费不同 partition。
即：同一订阅下，每个 partition 在任意时刻只会被一个消费者处理，其他消费者处于 standby 状态。
和Exclusive的区别就是Exclusive的消费者如果挂了就没有消费者了，但Failover就会有备消费者来消费。
3. Shared	多个消费者并发消费同一个订阅（支持并发）
4. Key_Shared	按 key 保证顺序，支持并发消费

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
1. 批量ACK合并(ack grouping)：
Pulsar 提供了 ack grouping 参数，允许客户端：
定时（如 100ms）或每 N 条消息后批量发送 ack
Broker 接收到的是一个“ack range”或“ack bitmap”而不是单条 ack。
2. 并发ACK时可能同时修改同一个分区的 ack 状态的数据一致性问题：
解决方案：
基于分片锁 + 高效 bitmap + 异步 flush
##### 分片锁？
1. 内部使用 fine-grained lock（分段锁）保护 Cursor 状态更新
2. ack 操作通过事件队列或线程池串行处理，避免竞争


##### bitmap？
Pulsar 使用 RoaringBitmap 存储已 ack 的消息 ID（entryId）。
每个 Cursor 内部维护一个 bitmap，表示「哪些 entry 被 ack」。
并发 ack 操作直接对 bitmap 做线程安全的更新（或通过事件线程串行执行）。

##### 使用事件线程（EventLoop）串行化 Cursor 操作？
每个 Topic/Subscription 在 Broker 侧有一个 Dispatcher（消息投递器）。
Dispatcher 绑定一个专属事件线程（单线程），负责处理所有 ack、redeliver、flow 控制等事件。
所以 ack 请求虽然来自多个 Consumer，但最终在 Broker 端被串行处理。


#### IO问题
1. ack操作放在内存中，ack请求不会立刻写磁盘，而是异步更新，后台线程会定期把 ack 状态同步到 BookKeeper（Pulsar 使用 Apache BookKeeper 存储消息日志（data）和 ledger metadata（ack、cursor 等））（持久化）:
(1)每个订阅（subscription）会维护一个 AckedMessageTracker 和一个 Cursor。
(2)Cursor 的更新是顺序追加（append-only），而不是随机写。并且写入是异步、批量的（由 Broker 定时 flush）。

###### AckedMessageTracker

###### Cursor


#### 消费者组 rebalance / failover 时的状态一致性
Pulsar 为每个 Subscription + Topic 创建一个 Cursor，Cursor 中记录：
(1)cumulative ack position（累积 ack 到哪）
(2)individual ack set（哪些消息单独 ack 了）
(3)pending ack（未 ack 的消息）

当消费者断开或 failover 时：
新消费者接管 subscription，Broker 会从 Cursor 中获取未 ack 消息，未 ack 消息支持 redelivery。


# ledger


