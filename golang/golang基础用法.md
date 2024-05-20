

<!-- toc -->

- [1. slice 原理：](#1-slice-%E5%8E%9F%E7%90%86)
- [2. 深度比较](#2-%E6%B7%B1%E5%BA%A6%E6%AF%94%E8%BE%83)
- [3. 强校验我们实现了接口的小方法：](#3-%E5%BC%BA%E6%A0%A1%E9%AA%8C%E6%88%91%E4%BB%AC%E5%AE%9E%E7%8E%B0%E4%BA%86%E6%8E%A5%E5%8F%A3%E7%9A%84%E5%B0%8F%E6%96%B9%E6%B3%95)
- [11. 接口和组合](#11-%E6%8E%A5%E5%8F%A3%E5%92%8C%E7%BB%84%E5%90%88)
- [12.Map、Reduce、Filter](#12mapreducefilter)
- [13.无缓冲 channel 和有缓冲 channel 的使用场景](#13%E6%97%A0%E7%BC%93%E5%86%B2-channel-%E5%92%8C%E6%9C%89%E7%BC%93%E5%86%B2-channel-%E7%9A%84%E4%BD%BF%E7%94%A8%E5%9C%BA%E6%99%AF)

<!-- tocstop -->

## 1. slice 原理：

```
type slice struct {
    array unsafe.Pointer //指向存放数据的数组指针
    len   int            //长度有多大
    cap   int            //容量有多大
}
```

三索引切片用法：

```
func main() {
    path := []byte("AAAA/BBBBBBBBB")
    sepIndex := bytes.IndexByte(path, '/')

    dir1 := path[:sepIndex:sepIndex]
    dir2 := path[sepIndex+1:]

    fmt.Println("dir1 =>", string(dir1)) //prints: dir1 => AAAA
    fmt.Println("dir2 =>", string(dir2)) //prints: dir2 => BBBBBBBBB

    dir1 = append(dir1, "suffix"...)

    fmt.Println("dir1 =>", string(dir1)) //prints: dir1 => AAAAsuffix
    fmt.Println("dir2 =>", string(dir2)) //prints: dir2 => BBBBBBBBB
}
```

这里面 dir1 的初始化，使用了三索引切片的用法 a[start:end:cap],它指定了新的切片的容量为 cap-start(示例中即是 sepIndex-0)，因此对 dir1 的 append 必定造成重新分配内存，因此底层数组将被复制一份放到新的内存中，从而对 dir 的 append 不会影响 dir2。

## 2. 深度比较

当我们复制一个对象时，这个对象可以是内建数据类型、数组、结构体、Map……在复制结构体的时候，如果我们需要比较两个结构体中的数据是否相同，就要使用深度比较，而不只是简单地做浅度比较。这里需要使用到反射 reflect.DeepEqual(),示例：

```
func main() {
    m1 := map[string]string{"one": "a","two": "b"}
    m2 := map[string]string{"two": "b", "one": "a"}
    fmt.Println("m1 == m2:",reflect.DeepEqual(m1, m2))
    //prints: m1 == m2: true

    s1 := []int{1, 2, 3}
    s2 := []int{1, 2, 3}
    fmt.Println("s1 == s2:",reflect.DeepEqual(s1, s2))
    //prints: s1 == s2: true
}
```

## 3. 强校验我们实现了接口的小方法：

```
type Shape interface {
    Sides() int
    Area() int
}
type Square struct {
    len int
}
func (s* Square) Sides() int {
    return 4
}
func main() {
    s := Square{len: 5}
    fmt.Printf("%d\n",s.Sides())
}

```

这个例子中，Square 没有实现 Area() int，导致他不是 Shape 接口。Go 语言编程圈里，有一个比较标准的做法：

```
var _ Shape = (*Square)(nil)

```

声明一个 \_ 变量（没人用）会把一个 nil 的空指针从 Square 转成 Shape，这样，如果没有实现完相关的接口方法，编译器就会报错。从而达到了强校验的目的。

4. 使用 StringBuffer 或是 StringBuild 来拼接字符串，性能会比使用 + 或 +=高三到四个数量级。

5. 避免在热代码中进行内存分配，这样会导致 gc 很忙。尽可能使用 sync.Pool 来重用对象。

6. 使用 lock-free 的操作，避免使用 mutex，尽可能使用 sync/Atomic 包（关于无锁编程的相关话题，可参看《无锁队列实现》或《无锁 Hashmap 实现》）。

7. 使用 I/O 缓冲，I/O 是个非常非常慢的操作，使用 bufio.NewWrite() 和 bufio.NewReader() 可以带来更高的性能。

8. 对于在 for-loop 里的固定的正则表达式，一定要使用 regexp.Compile() 编译正则表达式。性能会提升两个数量级。

9. 如果你需要更高性能的协议，就要考虑使用 protobuf 或 msgp 而不是 JSON，因为 JSON 的序列化和反序列化里使用了反射。

10. 你在使用 Map 的时候，使用整型的 key 会比字符串的要快，因为整型比较比字符串比较要快。

## 11. 接口和组合

```
type Widget struct {
    X, Y int
}

type Label struct {
    Widget        // Embedding (delegation)
    Text   string // Aggregation
}

type Button struct {
    Label // Embedding (delegation)
}

func (label Label) Paint() {
  fmt.Printf("%p:Label.Paint(%q)\n", &label, label.Text)
}

//因为这个接口可以通过 Label 的嵌入带到新的结构体，
//所以，可以在 Button 中重载这个接口方法
func (button Button) Paint() { // Override
    fmt.Printf("Button.Paint(%s)\n", button.Text)
}
func (button Button) Click() {
    fmt.Printf("Button.Click(%s)\n", button.Text)
}


```

Button.Paint() 接口可以通过 Label 的嵌入带到新的结构体，如果 Button.Paint() 不实现的话，会调用 Label.Paint() ，所以，在 Button 中声明 Paint() 方法，相当于 Override。

## 12.Map、Reduce、Filter

(1) Map 就是对数组的每个元素进行处理，处理结果还是数组
(2) Reduce 就是对数组的每个元素进行处理，处理结果是对所有元素参数计算后的单个结果值（比如对数组元素求和）
(3) Filter 就是过滤掉数组中其中一些元素，返回过滤后的数组

这三种操作写代码时，具体操作都应该是一个处理函数作为入参，将控制逻辑和业务逻辑分开。

## 13.无缓冲 channel 和有缓冲 channel 的使用场景

无缓冲 channel 的特点和使用场景
(1) 同步通信： 无缓冲 channel 保证发送和接收是同步进行的，即发送操作会阻塞，直到另一个 goroutine 在该 channel 上执行接收操作，这可以用于不同 goroutine 之间的同步。
(2) 确保交付： 使用无缓冲 channel 可以确保消息被直接交付给接收者，因为发送者会阻塞直到接收者准备好。
(3) 顺序保证： 无缓冲 channel 在并发程序中用于保证操作的执行顺序。

有缓冲 channel 的特点和使用场景：
(1) 异步通信： 有缓冲 channel 允许发送操作在缓冲区未满时不阻塞，这可以用于减少等待时间和提高并发性能。
(2) 流量控制： 有缓冲 channel 可以作为一个队列使用，对于生产者和消费者速率不一致的情况提供了缓冲。
(3) 并发任务限制·： 有缓冲 channel 可以用来限制处理并发任务的数量，例如通过限制缓冲区大小来控制同时运行的 goroutine 数量。
总的来说，无缓冲 channel 在需要确保 goroutines 之间同步执行时非常有用，而有缓冲 channel 则适用于需要某种程度的异步处理和流量控制的场景。
