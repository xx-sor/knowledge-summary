1.对象编程方法的黄金法则——“Program to an interface not an implementation”。
示例：
```
type Country struct {
    Name string
}

type City struct {
    Name string
}

type Stringable interface {
    ToString() string
}
func (c Country) ToString() string {
    return "Country = " + c.Name
}
func (c City) ToString() string{
    return "City = " + c.Name
}

func PrintStr(p Stringable) {
    fmt.Println(p.ToString())
}

d1 := Country {"USA"}
d2 := City{"Los Angeles"}
PrintStr(d1)
PrintStr(d2)

```

我们使用了一个叫Stringable 的接口，我们用这个接口把“业务类型” Country 和 City 和“控制逻辑” Print() 给解耦了。于是，只要实现了Stringable 接口，都可以传给 PrintStr() 来使用。
在go的标准库中有很多使用这种编程模式的示例。最著名的就是 io.Read 和 ioutil.ReadAll 的玩法，其中 io.Read 是一个接口，你需要实现它的一个 Read(p []byte) (n int, err error) 接口方法，只要满足这个规则，就可以被 ioutil.ReadAll这个方法所使用：
```
type Reader interface {
	Read(p []byte) (n int, err error)
}

// ReadAll reads from r until an error or EOF and returns the data it read.
// A successful call returns err == nil, not err == EOF. Because ReadAll is
// defined to read from src until EOF, it does not treat an EOF from Read
// as an error to be reported.
//
// As of Go 1.16, this function simply calls io.ReadAll.
func ReadAll(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}

// ReadAll reads from r until an error or EOF and returns the data it read.
// A successful call returns err == nil, not err == EOF. Because ReadAll is
// defined to read from src until EOF, it does not treat an EOF from Read
// as an error to be reported.
func ReadAll(r Reader) ([]byte, error) {
	b := make([]byte, 0, 512)
	for {
		if len(b) == cap(b) {
			// Add more capacity (let append pick how much).
			b = append(b, 0)[:len(b)]
		}
		n, err := r.Read(b[len(b):cap(b)])
		b = b[:len(b)+n]
		if err != nil {
			if err == EOF {
				err = nil
			}
			return b, err
		}
	}
}
```

2.