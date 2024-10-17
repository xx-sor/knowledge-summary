

<!-- toc -->

- [1. GC](#1-gc)
  * [三色标记法](#%E4%B8%89%E8%89%B2%E6%A0%87%E8%AE%B0%E6%B3%95)
  * [gc 过程](#gc-%E8%BF%87%E7%A8%8B)
    + [写屏障](#%E5%86%99%E5%B1%8F%E9%9A%9C)
    + [垃圾回收阶段 ：](#%E5%9E%83%E5%9C%BE%E5%9B%9E%E6%94%B6%E9%98%B6%E6%AE%B5-)
- [2.内存结构](#2%E5%86%85%E5%AD%98%E7%BB%93%E6%9E%84)
  * [构成](#%E6%9E%84%E6%88%90)
  * [内存分配策略](#%E5%86%85%E5%AD%98%E5%88%86%E9%85%8D%E7%AD%96%E7%95%A5)
- [3.channel](#3channel)
  * [理念](#%E7%90%86%E5%BF%B5)
  * [原理](#%E5%8E%9F%E7%90%86)
    + [channel 数据结构：](#channel-%E6%95%B0%E6%8D%AE%E7%BB%93%E6%9E%84)
    + [创建策略](#%E5%88%9B%E5%BB%BA%E7%AD%96%E7%95%A5)
    + [channel 发送数据原理 （ch<- i）：](#channel-%E5%8F%91%E9%80%81%E6%95%B0%E6%8D%AE%E5%8E%9F%E7%90%86-ch--i)
    + [channel 接收数据](#channel-%E6%8E%A5%E6%94%B6%E6%95%B0%E6%8D%AE)
    + [channel 实现信号传递](#channel-%E5%AE%9E%E7%8E%B0%E4%BF%A1%E5%8F%B7%E4%BC%A0%E9%80%92)
    + [channel的垃圾回收](#channel%E7%9A%84%E5%9E%83%E5%9C%BE%E5%9B%9E%E6%94%B6)
- [4.如何实现并发安全地读/写某个变量/资源](#4%E5%A6%82%E4%BD%95%E5%AE%9E%E7%8E%B0%E5%B9%B6%E5%8F%91%E5%AE%89%E5%85%A8%E5%9C%B0%E8%AF%BB%E5%86%99%E6%9F%90%E4%B8%AA%E5%8F%98%E9%87%8F%E8%B5%84%E6%BA%90)
  * [sync.Mutex](#syncmutex)
  * [使用一些库提供的原子操作的数据结构](#%E4%BD%BF%E7%94%A8%E4%B8%80%E4%BA%9B%E5%BA%93%E6%8F%90%E4%BE%9B%E7%9A%84%E5%8E%9F%E5%AD%90%E6%93%8D%E4%BD%9C%E7%9A%84%E6%95%B0%E6%8D%AE%E7%BB%93%E6%9E%84)
  * [使用 channel](#%E4%BD%BF%E7%94%A8-channel)
    + [sync.WaitGroup](#syncwaitgroup)
      - [方法](#%E6%96%B9%E6%B3%95)
      - [作用](#%E4%BD%9C%E7%94%A8)
- [5.变量可见性](#5%E5%8F%98%E9%87%8F%E5%8F%AF%E8%A7%81%E6%80%A7)
- [6.slice是并发安全的么？](#6slice%E6%98%AF%E5%B9%B6%E5%8F%91%E5%AE%89%E5%85%A8%E7%9A%84%E4%B9%88)
  * [具体原因](#%E5%85%B7%E4%BD%93%E5%8E%9F%E5%9B%A0)
  * [解决方法](#%E8%A7%A3%E5%86%B3%E6%96%B9%E6%B3%95)
  * [append的实现原理](#append%E7%9A%84%E5%AE%9E%E7%8E%B0%E5%8E%9F%E7%90%86)
- [7.竟态条件](#7%E7%AB%9F%E6%80%81%E6%9D%A1%E4%BB%B6)
  * [定义](#%E5%AE%9A%E4%B9%89)
- [8.string是并发安全的么？](#8string%E6%98%AF%E5%B9%B6%E5%8F%91%E5%AE%89%E5%85%A8%E7%9A%84%E4%B9%88)
  * [string源码](#string%E6%BA%90%E7%A0%81)
  * [代码说明](#%E4%BB%A3%E7%A0%81%E8%AF%B4%E6%98%8E)
- [9.unsafe.Pointer](#9unsafepointer)
  * [特性](#%E7%89%B9%E6%80%A7)
  * [应用](#%E5%BA%94%E7%94%A8)
  * [注意](#%E6%B3%A8%E6%84%8F)
- [锁](#%E9%94%81)

<!-- tocstop -->

## 1. GC

### 三色标记法

将程序中的对象分成白色 黑色 和灰色三类：

白色：潜在的垃圾，可能会被回收

黑色：活跃的对象，不会被回收

灰色：活跃的对象，有指向白色对象的指针

### gc 过程

开始垃圾回收时，不存在任何的黑色对象，会把根对象(不需要其他对象就可以访问到的对象：全局对象 栈对象)标记成灰色，垃圾回收只会从灰色的集合中取出对象开始扫描，当没有一个灰对象时标记阶段结束。

具体的扫描逻辑是：
(1) 从灰对象集合中选择一个灰色并标记成黑色；将黑对象指的所有对象都标记成灰色，来保证不会被回收，然后重复直到灰对象集合中没有灰对象
(2) 然后清理所有的白对象但是垃圾标记和正常程序是同时进行，所以有可能出现标记错的情况，比如扫描了 a 以及 a 所有的子节点后，这时候用户建立了 a 指向 b 的引用，这时 b 是白色会被回收，所以引入了屏障。它可以在执行内存相关操作时遵循特定的约束，在内存屏障执行前的操作一定会先于内存屏障后执行的操作。屏障有两种，写屏障和读屏障，因为读屏障需要在读操作中加入代码，对性能影响大，所以一般都是写屏障。

#### 写屏障

业界有两种写屏障 ： 插入写屏障和删除写屏障 1.7 用的插入写屏障 1.8 用的混合写屏障
(1) 插入写屏障： 当 A 对象从 A 指向 B 改成从 A 指向 C 时，把 BC 都改成灰色。
(2) 删除写屏障：在老对象的引用被删除时，将白色的老对象改成灰色
(3) 混合写屏障 ：将被覆盖的对象标记成灰色 & 没有被扫描的新对象也被标记成灰色 & 将创建的新对象都标记成黑色
(屏障必须遵守三色不变性 ： 强三色不变性:黑色对象不会指向白对象，只会指向灰色对象或者黑色对象 弱三色不变性：黑色对象指向的白色对象必须包含一条从灰色对象经由的多个白色对象的可达路径)

#### 垃圾回收阶段 ：

stw
开启写屏障
stw 结束
扫描根对象
依次处理灰对象
关闭写屏障
清理所有白对象
垃圾回收触发条件
用户触发 runtime.gc
堆内存比上次垃圾回收增长 100%
离上次垃圾回收超过 2min

参考文档 : [内存分配](https://draveness.me/golang/docs/part3-runtime/ch07-memory/golang-memory-allocator/)

## 2.内存结构

内存管理的基本单元是 mspan，他管理着一连串的页(8K),他会组成一个双向链表

### 构成

线程缓存(Thread Cache) ：属于每一个独立的线程，没有多线程，所以没有锁竞争，当线程缓存内存不够时，会使用中心缓存的内存。

中心缓存(Central Cache) : 这个需要互斥锁，他包含两个 spanSet，用来存储包含空闲单元和不包含空闲单元的 mspan，线程缓存从中心缓存获取内存的顺序是：清理过的，包含空闲空间的 spanSet，没有被清理过的，有空闲空间的 spanSet, 都没有找到从堆中申请新内存。

堆(Heap) ： 包含全局的中心缓存的列表 central 和 管理堆区的内存区域的 arenas，还有两颗二叉排序树 free 和 scav , free 存放空闲非垃圾回收 span(HeapIdle)， scav 存放空闲已垃圾回收 span。

### 内存分配策略

根据需要分配的内存大小选择不同的处理策略，根据对象的大小将对象分为微对象(0-16B),小对象(16B-32KB),大对象(>32KB)

微对象:先使用微型分配器，依次尝试线程缓存，中心缓存和堆 来分配内存。微型分配器可以将多个较小的内存分配请求合入一个内存块里，当内存块里所有的对象都要被回收时，整个内存块才能被回收。

小对象:依次尝试线程缓存，中心缓存和堆 来分配内存。确定分配对象的大小以及 spanClass(有 67 种，每一种规定了特定大小，mspan 的个数)

大对象: 直接在堆上分配内存，计算该对象需要的页数，按照一页(8K)的倍数在堆上申请内存

## 3.channel

### 理念
不要通过共享内存的方式进行通信，而是应该通过通信的方式共享内存


### 原理
#### channel 数据结构：
```
type hchan struct {
	qcount   uint    // Channel 中的元素个数；当sendx == recvx时，他可以用来判断是满了还是空了，避免二义性
	dataqsiz uint    // Channel 中的循环队列的长度
	buf      unsafe.Pointer    // Channel 的缓冲区数据指针
	elemsize uint16    // 当前chan中存的元素的类型的大小
	closed   uint32    // 通道是否被关闭的标志
	elemtype *_type    // 队列元素的类型
	sendx    uint    // Channel 的发送操作处理到的位置,
	recvx    uint    // Channel 的接收操作处理到的位置
	recvq    waitq    // 由于缓冲区不足而阻塞的接收goroutine的列表，使用双向链表 runtime.waitq 表示
	sendq    waitq    // 由于缓冲区不足而阻塞的发送goroutine的列表，使用双向链表 runtime.waitq 表示

	lock mutex   // 读写都用这一把锁，也就是说读的时候就不能写
}

// runtime.sudog 表示一个在等待列表中的 Goroutine，该结构中存储了两个分别指向前后 runtime.sudog 的指针以构成链表。
type waitq struct {
	first *sudog
	last  *sudog
}

//src/runtime/chan.go
type sudog struct{
    g *g //记录哪个协程在等待
    
    
    next *sudog
    prev *sudog
    elem unsafe.Pointer // 等待发送/接收的数据在哪里
    
    ...
    
    c *chan //等待的是哪个channel
}

```

#### 创建策略
无缓冲的直接分配内存
有缓冲的不包含指针，为hchan和底层数组分配连续的地址
有缓冲的channel且包含元素指针，会为hchan和底层数组分配地址

#### channel 发送数据原理 （ch<- i）：

(1)加锁(如 channel 已关闭会报错)
(2)当存在等待接收者时，会直接发送。直接发送过程是 ：将发送的数据直接**拷贝**到接收变量的内存地址上； 把等待接收数据的 goroutine 设置成 grunnable(可运行状态)，并把该 g 放到发送方所在的处理器(P，GMP模型的P)的 runnext 上等待执行，该处理器在下一次调度时会立刻唤醒数据的接收方
(3)当缓冲区存在空余空间时，将发送的数据写入 channel 的缓冲区(sendx)
(4)当不存在缓冲区或者缓冲区已满的时候，等待其他 goroutine 从 channel 接收数据
(5)解锁（4中挂起的情况也会解锁）

注：
runnext 是处理器（P）数据结构中的一个字段，它用于存储一个特定的 goroutine，这个 goroutine 将会在下一次该处理器进行调度时优先执行。这是一种优化机制，允许运行时系统快速响应某些事件，比如 channel 的通信。

#### channel 接收数据
(1)加锁。
(2)如果channel的写等待队列存在发送者goroutinue：
如果是无缓冲channel，直接从第一个发送者goroutinue将数据拷贝给接收变量，唤醒发送的goroutinue
如果是有缓冲channel（已满），将循环数组buf的队首元素拷贝给接收变量，将第一个发送者goroutinue的数据拷贝到buf循环数组，唤醒发送的goroutinue
如果channel的写等待不存在发送者goroutinue：
如果循环数组buf非空，将循环数组buf的队首元素拷贝给接收变量
如果循环数组buf为空，这个时候就会走阻塞接收的流程，将当前 goroutine 加入读等待队列，并挂起等待唤醒
(3)解锁

#### channel 实现信号传递

```
// struct{} 是一种零字节的结构体类型，通常用于信号传递，因为它不占用任何内存空间。
done := make(chan struct{})

...
// 代码的另外某处
// 这行代码阻塞当前 goroutine，直到从 done channel 中接收到空struct的信号。这个信号表示前置操作已经完成。
<-done

```

#### channel的垃圾回收
Golang 的垃圾回收机制对于 Channel 也适用。如果**一个 Channel 不再被任何 Goroutine 使用**，那么它所占用的内存空间就可以被回收。Golang 的垃圾回收是自动进行的，不需要程序员手动操作。


## 4.如何实现并发安全地读/写某个变量/资源

### sync.Mutex

```
mu := sync.Mutex{}
	mu.Lock()
	defer mu.Unlock()
	/// xxx 操作数据逻辑
```

### 使用一些库提供的原子操作的数据结构

(1) sync/atomic 包提供了一些底层的原子操作，可以用于实现简单的并发安全操作，如计数器、自增、自减等。

```
import (
    "sync/atomic"
)

var counter int64

func increment() {
    atomic.AddInt64(&counter, 1)
}

func getCounter() int64 {
    return atomic.LoadInt64(&counter)
}
```

(2) map 的话可以用 sync.Map

```
import (
    "sync"
)

var syncMap sync.Map

func storeData(key, value interface{}) {
    syncMap.Store(key, value)
}

func loadData(key interface{}) (interface{}, bool) {
    return syncMap.Load(key)
}
```

### 使用 channel

```
package main

import (
    "fmt"
    "sync"
)

// 定义一个操作类型
type Operation struct {
    action func()
    // 一个常用的用法，一个channel(下面的opChan)用来保证临界区的串行执行，另一个channel(done)用来传递执行完成的信号
    done   chan struct{}
}

func main() {
    var data int
    opChan := make(chan Operation)

    // 用来保证goroutine执行完，主程序才结束
    var wg sync.WaitGroup

    // 启动一个 goroutine 处理所有操作
    go func() {
        // 这是一种常见的用法，把函数传入channel，另起一个一直运行的goroutine，for range顺序执行，保证不出现并发安全问题
        for op := range opChan {
            op.action()
            close(op.done)
        }
    }()

    // 并发写操作
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(i int) {
            defer wg.Done()

            // struct{} 是一种零字节的结构体类型，通常用于信号传递，因为它不占用任何内存空间。
            done := make(chan struct{})
            opChan <- Operation{
                action: func() {
                    data += i
                },
                done: done,
            }
            // 这个语句是为了保证本次传入opChan的action执行完成，完成了才能执行defer中的wg.Done()；这样才能保证所有的goroutine执行完成后，main函数再结束
            <-done
        }(i)
    }

    // 等待所有写操作完成
    wg.Wait()
    close(opChan)

    fmt.Println("Final data value:", data)
}
```

#### sync.WaitGroup

##### 方法

Add(delta int)：增加或减少等待的 goroutine 计数。delta 可以是正数（增加计数）或负数（减少计数）。
Done()：减少等待的 goroutine 计数，相当于 Add(-1)。
Wait()：阻塞当前 goroutine，直到 WaitGroup 的计数器变为零。

##### 作用

(1) 等待一组 goroutine 完成：这是 WaitGroup 最常见的用途。你可以启动多个 goroutine，并使用 WaitGroup 来等待它们全部完成。

```
package main

import (
	"fmt"
	"sync"
)

func worker(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("Worker %d starting\n", id)
	// 模拟工作
	fmt.Printf("Worker %d done\n", id)
}

func main() {
	var wg sync.WaitGroup

	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go worker(i, &wg)
	}

	wg.Wait()
	fmt.Println("All workers done")
}

```

(2) 同步操作：在某些情况下，你可能需要确保某些操作在特定的顺序内完成。WaitGroup 可以帮助你实现这种同步。

```
// 在另一处执行传入opChan中匿名函数
...

// WaitGroup 被用来确保写入result操作在从 result channel 读取之前完成
func (this *LRUCache) Get(key int) int {
	result := make(chan int)
	this.wg.Add(1)
	this.opChan <- func() {
		curNode, ok := this.valMap[key]
		if !ok {
			result <- -1
		} else {
			this.updateRecentUse(curNode)
			result <- curNode.Val
		}
		close(result)
		this.wg.Done()
	}
	this.wg.Wait()
	return <-result
}

```

## 5.变量可见性

在 Go 语言中，标识符（如变量、函数、类型等）的可见性是通过首字母大小写来控制的。首字母大写的标识符是导出的（即公共的），可以被其他包访问；首字母小写的标识符是未导出的（即私有的），只能在定义**它们的包内**访问。

Go 语言并没有像 C++ 那样的 private、protected 等访问控制修饰符。Go 语言的设计哲学是简单和明确，因此它只提供了包级别的可见性控制。
如果想让一个结构体的成员变量只有这个结构体的实例自己能访问，一种方法是将结构体定义在一个单独的包中，并将成员变量定义为未导出。这样，只有这个包内的代码可以访问这些成员变量，而包外的代码只能通过导出的方法来访问。
如果想让一个结构体的成员变量只有这个结构体的实例自己能访问，可以把这个结构体放到单独的包。

## 6.slice是并发安全的么？
很显然slice不是，因为他的实现是：
```
type slice struct {
    ptr unsafe.Pointer
    len int
    cap int
}
```
这个实现中连锁都没有，不可能是并发安全的。

### 具体原因
(1)共享底层数组：切片是对数组的封装，它包含指向数组的指针、切片的长度和容量。当多个切片共享同一个底层数组时，如果没有适当的同步，一个goroutine对底层数组的修改可能会影响到其他goroutine。

(2)切片操作非原子性：切片的操作，如append、slice[1:4]等，不是原子性的。这意味着在操作过程中，其他goroutine可能会看到不一致的状态。

(3)长度和容量的修改：当你对切片进行append操作时，如果切片的容量不足以容纳更多的元素，Go运行时会分配一个新的底层数组，并更新切片的指针、长度和容量。如果这个时候有另一个goroutine正在读取或写入切片，就可能会发生竞态条件。


### 解决方法
(1)加锁

(2)channel
```
package main

import (
    "fmt"
)

func main() {
    ch := make(chan int, 5)
    done := make(chan bool)

    for i := 0; i < 5; i++ {
        go func(i int) {
            ch <- i
        }(i)
    }

    go func() {
        for i := 0; i < 5; i++ {
            fmt.Println(<-ch)
        }
        done <- true
    }()

    <-done
}
```

### append的实现原理
当调用 append 函数时，它会首先检查切片的容量是否足够。如果容量足够，append 函数会在底层数组的未尾添加新的元素，并更新切片的长度。如果容量不足，append 函数会创建一个新的底层数组，并将原始切片的元素复制到新的底层数组中，然后再添加新的元素。其中更新容量时，如果需要重新分配，当容量小于1024时，新的数组容量翻倍；否则变为1.25倍。前者是为了防止频繁申请内存；后者是为了防止浪费内存。


## 7.竟态条件
### 定义
竞态条件是并发编程中的一个概念，它发生在两个或多个操作必须以正确的顺序执行，但程序的运行时行为却无法保证这一顺序，从而导致不可预知的结果。简单来说，当多个线程或进程访问和修改同一数据时，最终的结果依赖于这些线程或进程的具体执行顺序，这就是竞态条件。


## 8.string是并发安全的么？
golang中string是读并发安全，但写非并发安全的。
string一旦被创建，便不能修改。我们代码中的修改，实际上都是重新申请了一片新的内存。

### string源码
```
type stringStruct struct {
	str unsafe.Pointer
	len int
}
```

### 代码说明
(1)并发读取字符串是安全的：

```go
func main() {
    s := "hello, world"
    go func() {
        fmt.Println(s)
    }()
    go func() {
        fmt.Println(s)
    }()
}
```

在这个例子中，我们在两个 goroutine 中同时读取同一个字符串 `s`。这是完全安全的，因为字符串是不可变的，所以我们不需要担心其中一个 goroutine 会改变 `s` 的内容。

(2)并发修改字符串变量可能会导致竞态条件：

```go
func main() {
    s := "hello"
    go func() {
        s = s + ", world"
    }()
    go func() {
        s = s + "!"
    }()
    time.Sleep(time.Second)
    fmt.Println(s)
}
```

在这个例子中，我们在两个 goroutine 中同时修改同一个字符串变量 `s`。这可能会导致竞态条件，因为我们不能预测 `s` 的最终值是什么，这取决于哪个 goroutine 最后执行了赋值操作。

(3)使用锁来同步并发修改字符串变量：

```go
func main() {
    var mu sync.Mutex
    s := "hello"
    go func() {
        mu.Lock()
        s = s + ", world"
        mu.Unlock()
    }()
    go func() {
        mu.Lock()
        s = s + "!"
        mu.Unlock()
    }()
    time.Sleep(time.Second)
    fmt.Println(s)
}
```

在这个例子中，我们使用了一个互斥锁来同步两个 goroutine 对字符串变量 `s` 的修改。这样我们就可以确保 `s` 的最终值是 "hello, world!"，因为每次只有一个 goroutine 可以修改 `s`。

## 9.unsafe.Pointer
`unsafe.Pointer` 是 Go 语言中的一个特殊类型，它允许程序员在类型系统的约束下进行某些操作。它的定义如下：
```go
type Pointer *ArbitraryType
```

### 特性
(1)任何类型的指针都可以被转化为 `unsafe.Pointer`。
(2)`unsafe.Pointer` 可以被转化为任何类型的指针。
(3)`unsafe.Pointer` 可以被转化为 `uintptr`（一个用于存储指针的整数类型），反之亦然。

### 应用
Go 语言中的 `unsafe` 包和 `unsafe.Pointer` 主要用于绕过 Go 语言的类型系统，进行一些底层的、灵活的操作。它们的使用场景包括：
(1)实现跨类型的转换：你可以将一个类型的指针转换为 `unsafe.Pointer`，然后再转换为另一个类型的指针。
```
package main

import (
	"fmt"
	"unsafe"
)

func main() {
	i := 10
	f := *(*float64)(unsafe.Pointer(&i))
	fmt.Println(f)
}
```

(2)访问结构体的私有字段：在 Go 语言中，你不能直接访问一个包外的结构体的私有字段。但是，你可以通过 `unsafe.Pointer` 来绕过这个限制。
```
package main

import (
	"fmt"
	"unsafe"
)

type Foo struct {
	a int
	b string
}

func main() {
	f := Foo{10, "hello"}
	p := (*string)(unsafe.Pointer(uintptr(unsafe.Pointer(&f)) + unsafe.Offsetof(f.b)))
	*p = "world"
	fmt.Println(f)
}
```

(3)实现二进制序列化：如果你知道一个数据结构的内存布局，你可以使用 `unsafe.Pointer` 来直接将其序列化为二进制数据，或者从二进制数据中反序列化。
```
package main

import (
	"fmt"
	"unsafe"
)

type Foo struct {
	a int
	b string
}

func main() {
	f := Foo{10, "hello"}
	p := (*[unsafe.Sizeof(f)]byte)(unsafe.Pointer(&f))
	fmt.Println(p)
}
```

(4)调用系统调用或者调用 C 语言的函数：在这些情况下，你可能需要直接操作内存或者构造特定的数据结构。
需要注意的是，虽然 `unsafe.Pointer` 提供了很大的灵活性，但是它也带来了一些风险。使用 `unsafe.Pointer` 可能会破坏类型安全，导致程序崩溃或者数据损坏。因此，除非必要，否则应该尽量避免使用 `unsafe.Pointer`。
```
package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"unsafe"
)

func main() {
	cs := C.CString("Hello from stdlib")
	fmt.Println(*(*string)(unsafe.Pointer(&cs)))
	C.free(unsafe.Pointer(cs))
}
```

### 注意
在实际的代码中应该尽量避免使用 unsafe.Pointer



map
使用拉链法解决 hash 碰撞问题

数据结构
hmap:可以理解是一个 hash 槽

count 当前哈希表中的元素数量 ；

B 表示当前哈希表持有的 bucket 的数量。但是因为哈希表中桶的数量都是 2 的倍数，所以该字段会存储对数，即 len(buckets) == 2^B

hash0 是哈希的种子，他能为 hash 函数的结果引入随机性，这个值在创建哈希表时确定，并在调用哈希函数时作为参数传入；

oldbuckets 是哈希在扩容时用于保存之前 buckets 的字段 ，它的大小是当前 buckets 的一半

buckets ： bmap 的 list，

bmap: buckets 中的值，每一个 bmap 都能存储 8 个键值对，当哈希表中存储的数据过多，的那个桶已经装满时就会使用 extra 中桶存储溢出的数据，这两种不同的桶在内存中连续存储的。数据结构主要包含一个简单的 tophash 结构，存储了键的 hash 的高 8 位，通过对比不同键的哈希的高 8 位可以减少访问键值对次数以提高性能。不过溢出桶只是临时方案，创建过多的溢出桶最终也会导致 hash 的扩容

创建 map
计算 hash 占用的内存是否溢出

获取一个随机的 hash 种子

计算需要的最小需要桶的数量

创建用于保存桶的数组，根据 B 计算出需要创建的桶数量并在内存中分配一段连续的空间用户存储数据，

读写,扩容操作
读取 : 通过 hash 表设置的 hash 函数和种子获取当前键对应的 hash，再拿到该键值对所在的桶序号(hash 最低几位)和 hash 高位的 8 位数字。然后会依次遍历正常桶和溢出桶的数据，先比较 hash 的高 8 位和桶中存储的 tophash，后比较传入的值和桶中的值以加速数据读写。用于选择桶序号的是 hash 的最低几位，用于加速访问的是 hash 高 8 位，这种设计避免同一个桶中有大量相等的 tophash,

写入： 先读取，如果键值对的 hash 不存在，会为新键值对规划存储的内存地址

扩容 ： 在一下两种情况会触发 hash 的扩容 1 装载因子 > 6.5 ; (装载因子 = 元素数量/桶数量 ) 2 hash 使用了太多溢出桶

如果这次扩容是溢出桶太多导致的，就是等量扩容 ，否则就是翻倍扩容 。

扩容
创建一组新桶和预创建的溢出桶，然后将原有的桶数组设置到 oldbuckets 上，将新的空桶设置到 buckets 上，溢出桶也用了相同逻辑。如果等量扩容，旧桶和新桶是 1 对 1 关系，当翻倍扩容时，每个旧桶元素会分流到新创建的 2 个桶中，比如扩容前桶号是 3(11) 扩容后分流到 3(011)和 7(111)。
当 hash 表处于扩容状态时，每次写入或删除都会触发增量拷贝

遍历 ： 会引入一个随机数来随机选择一个遍历桶的位置，会先选一个正常桶开始遍历，然后遍历所有的溢出桶，然后依次按照索引顺序遍历其他桶

sync map
golang map 是协程不安全的，sync.map 是协程安全的，采用读写分离的方式降低锁粒度，适用于读多写少的场景，对于写多的场景会导致 read map 缓存失效，需要加锁，导致冲突变多

数据结构
mu 互斥锁 ;

read 存储读的数据，只读，所以并发安全，每次读写的时候 golang 都会吧类型转换成 readOnly

readOnly 里面是一个 map 结构个一个标记和 drity 数据是否相同的字段，misses 计数用的，每次从 read 里读取失败 +1 ; drity 包含最新写入的数据，当 misses 打到一定值，将 dirty 赋值给 read

读取，存储，删除
读取 ： 读 read 表 如果没读到并且 drity 结果一样，就返回结果。否则就加锁，然后再读 read，如果还是没读到并且和 drity 结果不一样，就读 drity，然后 misses++，然后解锁。这里在要做一次判断是害怕之前的判断和加锁操作是非原子的。
中间这个 misses 的值如果比 dirty 的长度长，就会吧 drity 的值赋给 read ，drity 置为空，misses 置为 0

存储 ：如果存在并且没有标记成已删除，就直接返回，否则先查询 read，如果标记为以删除，就把值加入到 drity 中，更改指针的值。如果 read 里不存在的话，先看 dirty 和 read 里数据是否相同 如果相同就再判断 drity 是不是 nil，如果是 nil，就会遍历 read 中的值赋给 drity，并且把此时为 nil 的 key 置为已删除。然后再重置 read。 如果不同就把值放到 drity 里。整个这个过程会加锁。

删除：先读 read，如果读到，则把 read 的值置为空，如果没读到并且和 drity 值不一致就会加锁，然后再读 read，如果结果还是没读到且和 dirty 值不一致，就会删除 dirty 里的值，是吧 Entery 的值置为 nil，然后解锁。

select
让 goroutine 同时等待多个 channel 可读或可写，在多个 channel 状态改变前，select 会一直阻塞当前 goroutine。select 里的 case 中的表达式必须都是 channel 的收发操作

现象
1.select 能在 channel 上进行非阻塞的收发操作(利用 default)

2.select 在遇到多个 channel 同时相应时，会随机执行一种情况(为了避免饥饿问题)

实现原理
根据 select 语句情况优化语句
● select 不存在任何 case ： 直接阻塞
● select 只存在一个 case ： 编译器会改写成 if
● select 仅包含两个 case，其中一个是的 default ： 编译器认为是一次非阻塞收发操作
● 普通情况 ： 通过 selectgo 获取执行 case 的索引，并通过多个 if 执行对应 case 的代码

随机生成一个遍历的轮询顺序 pollOrder 并根据 channel 地址生成加锁顺序 lockOrder

根据 pollOrder 遍历所有的 case 看是否有立刻可以处理的 channel
如果存在，直接获取 case 对应的索引并返回
如果不存在，将当前 goroutine 加入到 channel 的收发队列，并挂起当前 Goroutine

当调度器唤醒当前 goroutine 时，按照 lockOrder 遍历所有 case，查找需要被处理的索引

defer
原理&处理方式
堆分配(早期，兜底)，栈分配（1.13，节省开销），开放编码(1.14)

堆分配： 编译期间将 defer 关键字转换成 deferproc 函数(负责创建新的延迟调用)，在调用 defer 函数的结尾插入 deferreturn 函数(负责在函数调用结束时执行所有的延迟调用),运行时调用 deferproc 会将一个新的\_defer 结构体(包括参数和结果的内存大小 ，栈指针和调用方程序计数器，defer 传入的函数等等)追加到当前 goroutine 的链表头，运行时调用 deferreturn 会从链表取出该结构体并依次执行

栈分配 ： defer 在函数体中最多执行一次时，会将 defer 结构体分配到栈上并调用

开放编码 ： 编译期间根据 defer 和 return 的个数判断是否开启开放编码优化，如果 defer 的执行可以在编译期间确定，会在函数返回前直接插入相应代码，否则由运行时的 deferreturn 处理

panic 和 recover
panic 能够改变程序的控制力，调用 panic 后会立刻停止执行当前函数的剩余代码，并在当前 goroutine 中递归执行调用方的 defer recover 可以中止 panic 造成的程序崩溃，他是一个只能在 defer 中发挥作用的函数

panic 原理
将 panic 和 recover 转换成 gopanic 和 gorecover 函数

运行过程中遇到 gopanic 方法，会将 goroutine 的链表中依次取出 \_defer 结构体并执行

如果执行延迟函数时遇到了 gorecover : 在这次调用结束后 gopanic 会从 \_defer 结构体中取出程序计数器和栈指针恢复程序；并跳回 deferproc，再跳回 deferreturn 并恢复正常流程

如果没有遇到 gorecover 就会依次遍历所有\_defer,并最后调用 fatalpanic 中止程序打印 panic 的参数并返回错误码

interface
空的 interface ： eface
\_type 字段：指向一个运行时类型信息的结构体 。 size，ptrdata 表示 interface 对象的类型信息，hash 哈希值，用于 map，
align 和 fieldalign 用与内存对齐，kind 类型的种类(bool int )，equal ： 判断是否相等 ，gcData ： 垃圾回收数据，
Data ：内存指针，指向 interface 实例对象信息的存储地址，可以获取对象的具体属性的信息

非空的 interface 数据结构 ： iface 。关键数据结构是
data ：同 eface
tab ： itab ： inter ： 指向接口类型本身信息的指针，\_type : 指向具体类型信息的指针 fun： 数组，指向实现接口的具体类型方法的指针

make 和 new
new 用于分配内存，会返回类型的指针，值会被初始化为”0“ make 仅用于分配和初始化 slice、map、channel 类型的对象，3 种类型都是结构，返回类型是结构不是指针。

## 锁
sync.Mutex
数据结构

state 互斥锁的状态

sema 控制锁状态的信号量组成，默认情况下，互斥锁的状态位都是 0 ，state 的 int32 中不同的位代表了不同的状态

mutexLocked 表示互斥锁的锁定状态

mutexWoken 表示是否有被唤醒的 goroutine

mutexStraving 当前互斥锁进入饥饿状态

waitersCount 当前互斥锁上等待的 goroutine 个数

饥饿模式是 1.9 引入的优化，正常情况下锁的等待者会按照先进先出的顺序获取锁，但是刚被唤起的 goroutine 与新创建的 goroutine 竞争时，大概率获取不到锁，为了减少这种情况，一旦 goroutine 超过 1ms 没有获取到锁，该 g 就会把锁切换到饥饿模式，在饥饿模式中，互斥锁会直接交给等待队列最前面的 goroutine，新的 goroutine 在该状态下只会在队列末尾等待，如果一个 goroutine 获得了互斥锁并且它在队列的末尾或者它等待时间<1ms,当前互斥锁就会切回正常状态

加锁过程 ：

如果互斥锁处于初始化状态，会通 mutexLocked 加锁,如果互斥锁处于 mutexLocked 状态并在普通模式下工作，会进入自旋，执行 30 次 PAUSE 指令消耗 cpu 时间等待锁的释放

如果当前 goroutine 等待锁的时间超过 1ms，互斥锁就会切换到饥饿模式，
互斥锁在正常情况下会尝试获取锁的 goroutine 切换至休眠状态，等待锁的持有者唤醒

如果当前 goroutine 是互斥锁上的最后一个等待的协程或者等待的时间小于 1ms，那么它会将互斥锁切回正常模式

解锁过程：

当互斥锁已经被解锁时，调用 unlock 会直接抛出异常

当互斥锁处于饥饿模式时，将锁的所有权交给队列中的下一个等待者，等待者会负责设置 mutexLocked 标志位

当互斥锁处于普通模式时，如果没有 goroutine 等待锁的释放或者已有被唤醒的 goroutine 获取了锁，会直接返回，否则会唤醒当前 goroutine

sync.RWMutex
数据结构

w (mutex)复用互斥锁的能力；

writerSem readerSem 用于写等待读和读等待写的信号量，readerCount 存储了当前正在执行的读操作数量，readerWait 表示当前写操作被阻塞时等待的读操作个数

获取写锁时(rwmutex.lock)

调用 mutex 的 lock 阻塞后续操作

给 readerCount - rwmutexMaxReaders（2^30）阻塞后续读操作 3 如果有其他 goroutine 持有互斥锁的读锁，该 g 会进入休眠等待所有读锁持有者执行结束后释放 writerSem 将其唤醒

写锁释放(rwmutex.ulock)

将 readerCount 变会正数，释放读锁

for 循环释放所有因为获取读锁而陷入等待的 goroutine

调用 mutex.unlock 释放写锁

获取读锁(rwmutex.rlock) ：
readerCount ++；

如果该值为负数，则其他 g 获取了写锁，该 g 就会陷入休眠等待锁的释放

非负数，则获取成功

释放读锁(rwmutex.RUnlock) : readerCount-- 1.如果返回值>=0 则解锁成功 2.如果<0，则说明有正在执行的写操作，则会减少 readerWait 并在所有读操作后释放触发写操作的信号量 writerSem，该信号量被触发后，会尝试唤醒尝试获取写锁的 g

WaitGroup
等待一组 goroutine 的返回，使用 waitgroup 将原本顺序执行的代码在多个 goroutine 中并发执行，加快速度 数据结构： noCopy 保证不会被开发者通过再赋值的方式拷贝；state1 存储状态和信号量

反射
反射机制就是通过来获取对象的类型信息或者结构信息，再进行访问和修改的能力。

反射 3 定律： 1.从 interface{}变量可以反射出反射的对象 2.从反射对象可以获取 interface{}变量 3.要修改反射对象，其值必须可设置

切片和数组
数组
是一种有固定长度的基本数据结构，一旦创建就不影响改变长度。数组是值拷贝传递。数组中的元素小于等于 4 个，所有的变量会在栈上初始化，否则会在静态存储区初始化，然后拷贝到栈上

切片
slice 本身是一个特殊的引用类型，自身是一个结构体，属性 len 表示可用元素数量，读写操作不能超过这个限制，cap 表示最大扩张容量。如果 slice 在 append 时容量超过 cap 会触发扩容 分配一个容量翻倍的内存 不再影响原有内存
切片初始化问题： 切片初始化的时候 如果不设置 cap cap 会为 0，如果之后频繁 append 会触发多次扩容，可以预先设置一个 cap 比如 1024

内存逃逸
在函数内部分配的变量，由于某些原因生命周期被延长，必须在堆上分配，而不是栈上。

分析逃逸：go build -gcflags -'m'

可能带来的影响 1.堆上分配内存慢 2.垃圾回收压 3.有可能内存泄露

内存逃逸可能情况 1.函数返回局部指针：函数返回一个局部变量的指针时，这个变量就会发生逃逸，因为局部变量生命周期本来应该在这个函数结束时结束，但是返回指针代表外部也能访问他
res 会逃逸到堆上

```
func Add(x,y int)*int{
	res := 0
	res =x+y
	return &res
}
```

动态分配逃逸 ： 当通过 make new 动态分配内存时，尤其是编译器无法确定大小时

```
funcmakeSlice(sizeint)[]int{
	s:=make([]int,size)
	return s
}
```

interface{} 类型逃逸 ： 当一个具体的变量被赋值给一个 interface 类型时，编译器没法确定具体类型

str:="aaaaa"
fmt.Println("%v",str)
闭包引用逃逸 ：闭包引用了他的外部函数的局部变量 这个局部变量本来应该在函数返回时结束，但是由于闭包，这个局部变量需要在函数执行完毕后还是可用

```
func a() func() int {
	x:= 100
    return func() int {
		x++
        returnx
    }
}
```

32.golang 中函数的编译：
先看 2 个例子：

```
func canFinish(numCourses int, prerequisites [][]int) bool {
    edges := make([][]int, numCourses)
    visited := make([]int, numCourses)
    result := []int{}

    // 初始化edge
    for _, v1 := range prerequisites {
        edges[v1[1]] = append(edges[v1[1]], v1[0])
    }

    dfs := func(curNode int) bool {
        visited[curNode] = 1
        for _, dstNode := range edges[curNode] {
            if visited[dstNode] == 0 {
                // 编译错误！！！报dfs未定义
                valid := dfs(dstNode)
                // 如果已经出现环了，不用继续了
                if !valid {
                    return valid
                }
            } else if visited[dstNode] == 1 {
                // 访问到还没回溯的节点，说明出现了依赖环
                return false
            }
            // 不需要==2的分支，因为不需要处理
        }
        // 标记此节点已完成
        visited[curNode] = 2
        // 加入拓扑排序的结果
        result = append(result, curNode)
        return true
    }

    // dfs
    for i, _ := range edges {
        if visited[i] == 0 {
            valid := dfs(i)
            // 出现环就直接返回了
            if !valid {
                return valid
            }
        }
    }

    return true
}
```

以上代码编译时就在 12 行报错。

```
func canFinish(numCourses int, prerequisites [][]int) bool {
    edges := make([][]int, numCourses)
    visited := make([]int, numCourses)
    result := []int{}

    // 初始化edge
    for _, v1 := range prerequisites {
        edges[v1[1]] = append(edges[v1[1]], v1[0])
    }

    // 先定义函数变量
    var dfs func(curNode int) bool
    dfs = func(curNode int) bool {
        visited[curNode] = 1
        for _, dstNode := range edges[curNode] {
            if visited[dstNode] == 0 {
                valid := dfs(dstNode)
                // 如果已经出现环了，不用继续了
                if !valid {
                    return valid
                }
            } else if visited[dstNode] == 1 {
                // 访问到还没回溯的节点，说明出现了依赖环
                return false
            }
            // 不需要==2的分支，因为不需要处理
        }
        // 标记此节点已完成
        visited[curNode] = 2
        // 加入拓扑排序的结果
        result = append(result, curNode)
        return true
    }

    // dfs
    for i, _ := range edges {
        if visited[i] == 0 {
            valid := dfs(i)
            // 出现环就直接返回了
            if !valid {
                return valid
            }
        }
    }

    return true
}
```

以上这段代码是正确的。
为什么需要在写递归函数的时候先声明？全局作用域中的递归函数显然是不需要声明的，例如：

```
func FindGroup(root *pb.GroupInfoV3, groupID string) *pb.GroupInfoV3 {
	if root == nil {
		return nil
	}
	if root.GroupId == groupID {
		return root
	}
	// BFS
	for _, v := range root.ChildGroupList {
		curRes := FindGroup(v, groupID)
		if curRes != nil {
			return curRes
		}
	}
	return nil
}
```

golang 编译函数的原理：
在 Go 语言中，编译器在处理源代码时会经过几个阶段，包括
(1)解析（parsing）
(2)类型检查（type checking）
(3)中间代码生成（intermediate code generation）
(4)优化（optimization）
(5)最终的机器代码生成（code generation）。
在解析阶段，编译器会构建出一个抽象语法树（AST），这个树结构表示了源代码的语法结构。
对于全局作用域中的函数声明，编译器在解析阶段就会将其加入到全局作用域中。这意味着一旦函数被解析，它就会被加入到全局的命名空间，之后的编译阶段都可以访问到这个函数。
全局递归函数不需要先声明是因为 Go 语言的编译器在解析代码时会进行两次遍历：
第一次遍历：在解析阶段中。编译器会收集所有的类型声明和函数声明，但不会深入函数体内部。这意味着所有的全局类型和函数在这个阶段都会被声明，因此它们在之后的任何地方都是可见的。
第二次遍历：通常在类型检查阶段中。编译器会处理函数体内部的代码，包括局部变量的声明和表达式的求值。在这个阶段，函数体内部的代码才会被实际解析。

局部作用域中的递归函数需要先声明的原因是，匿名函数是在函数体内部定义的，它们不会在第一次遍历时被处理。因此，如果你想在匿名函数内部递归调用它自己，你必须先声明一个变量来持有这个匿名函数，这样在匿名函数体内部就可以通过这个变量引用自己。
这种设计允许编译器在处理局部作用域时正确地处理变量的作用域和生命周期，同时也确保了代码的一致性和可读性。在局部作用域中，所有的变量和函数都必须在使用前声明，这样编译器才能正确地管理它们的作用域和生命周期。
