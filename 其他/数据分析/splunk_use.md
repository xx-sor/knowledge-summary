# 管理员手册

## 配置文件

### 编辑配置文件
⼤多数 Splunk 配置信息存储在 .conf ⽂件中。这些⽂件位于 Splunk 安装⽬录（在⽂档中通常称为 $SPLUNK_HOME）下的/etc/system 下。⼤多数情况下，可将这些⽂件复制到本地⽬录并使⽤⾸选的⽂件编辑器对其进⾏更改。

## 数据管道
数据在从原始输⼊转换为可搜索事件过程中经历了⼏个阶段。此过程称为数据管道，由以下四个阶段构成：
输⼊
分析
索引
搜索
![](images/image-2024-08-28-10-37-27.png)

## 开发应用
Splunk 不⽀持 Splunkbase 上的所有应⽤和加载项。
有关应⽤和加载项的⽀持选项列表，参阅 Splunk 开发⼈员⻔户中的“Splunkbase 上应⽤的⽀持类型”。
### 指导
dev.splunk.com

## kv store
主要存储元数据、配置、状态信息等，而不用来存储需要进行大规模搜索和分析的数据。kv store底层使用的是 MongoDB 来存储数据。

### 内置的 MongoDB
Splunk Enterprise 自带一个嵌入式的 MongoDB 实例，这个实例只用于 KV Store 的数据存储和操作。当安装 Splunk Enterprise 时，MongoDB 已经被包含在安装包中并自动配置好了。
在 Splunk Enterprise 的本地部署中，MongoDB 的数据文件通常存储在 $SPLUNK_HOME/var/lib/splunk/kvstore/mongo 目录下。
当启动 Splunk Enterprise 时，Splunk 会自动启动内置的 MongoDB 实例，并管理其配置和操作。用户不需要进行任何额外的设置或下载。

#### 安装目录
MongoDB 二进制文件： $SPLUNK_HOME/bin/splunkd
MongoDB 数据存储目录： $SPLUNK_HOME/var/lib/splunk/kvstore/mongo

# 数据
## 存储
存储在称为索引（Index）的结构中。这些索引实际上是存储在文件系统中的一组文件和目录，而没有使用传统的关系型数据库。

### 索引
数据存储在索引中，每个索引代表一个逻辑数据存储单元。
索引由一组事件数据和索引文件组成，索引文件用于加速数据检索。

#### 索引示例
假设有一个 Web 服务器的访问日志数据，你可以将这些日志数据导入 Splunk，并存储在一个名为 web_access 的索引中。
数据示例（Web 访问日志）：
```
127.0.0.1 - - [10/Oct/2023:13:55:36 -0700] "GET /index.html HTTP/1.1" 200 1024
192.168.1.1 - - [10/Oct/2023:13:56:01 -0700] "POST /login HTTP/1.1" 302 512
```
在 Splunk 中，数据会被分解成事件：
```
_event_1:
{
  "host": "127.0.0.1",
  "time": "10/Oct/2023:13:55:36 -0700",
  "method": "GET",
  "url": "/index.html",
  "status": 200,
  "bytes": 1024
}

_event_2:
{
  "host": "192.168.1.1",
  "time": "10/Oct/2023:13:56:01",
  "method": "POST",
  "url": "/login",
  "status": 302,
  "bytes": 512
}

```
查询示例：
```
index=web_access | stats count by status
```
该查询会搜索 web_access 索引中的所有事件，并按 HTTP 状态码统计事件的数量。



### 数据分类
热数据（Hot Data）：最近写入的数据，存储在内存中并实时可搜索。热数据存储在 hot 目录中。
温数据（Warm Data）：热数据冷却后转移到温数据，这些数据仍然可搜索。温数据存储在 warm 目录中。
冷数据（Cold Data）：较旧的数据，存储在冷存储中，但仍然可搜索。冷数据存储在 cold 目录中。
冻结数据（Frozen Data）：超过保留期的数据，默认情况下会被删除，可以配置为转移到外部存储。

### 存储位置
默认情况下，Splunk 的数据存储在 $SPLUNK_HOME/var/lib/splunk 目录下。在该目录下，每个索引都有一个单独的子目录。例如：
```
$SPLUNK_HOME/var/lib/splunk/
├── _internaldb
├── _audit
├── main
├── my_custom_index
└── ...
```
每个索引目录中包含热、温、冷数据的子目录。例如：
```
$SPLUNK_HOME/var/lib/splunk/main/
├── db
│   ├── hot_v1_0
│   ├── warm_v1_0
│   ├── ...
├── colddb
└── thaweddb

```


