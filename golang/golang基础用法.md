1. slice 原理：

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

2. 深度比较
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

3. 强校验我们实现了接口的小方法：

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

11. 接口和组合

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

12.Map、Reduce、Filter
(1) Map 就是对数组的每个元素进行处理，处理结果还是数组
(2) Reduce 就是对数组的每个元素进行处理，处理结果是对所有元素参数计算后的单个结果值（比如对数组元素求和）
(3) Filter 就是过滤掉数组中其中一些元素，返回过滤后的数组

这三种操作写代码时，具体操作都应该是一个处理函数作为入参，将控制逻辑和业务逻辑分开。
