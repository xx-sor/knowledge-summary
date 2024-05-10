## 1.对象编程方法的黄金法则——“Program to an interface not an implementation”。

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

我们使用了一个叫 Stringable 的接口，我们用这个接口把“业务类型” Country 和 City 和“控制逻辑” Print() 给解耦了。于是，只要实现了 Stringable 接口，都可以传给 PrintStr() 来使用。
在 go 的标准库中有很多使用这种编程模式的示例。最著名的就是 io.Read 和 ioutil.ReadAll 的玩法，其中 io.Read 是一个接口，你需要实现它的一个 Read(p []byte) (n int, err error) 接口方法，只要满足这个规则，就可以被 ioutil.ReadAll 这个方法所使用：

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

## 2.Functional Options

### 解决什么问题

在编程中，我们经常需要对一个对象（或是业务实体）进行相关的配置。其中有些是必填的，有些是有默认值，可以选填的。
如：

```
type Server struct {
    Addr     string    // 必填
    Port     int    // 必填
    Protocol string     // 选填
    Timeout  time.Duration  // 选填
    MaxConns int    // 选填
    TLS      *tls.Config    // 选填
}

```

针对这样的配置，我们需要有多种不同的创建不同配置 Server 的函数签名。

```
func NewDefaultServer(addr string, port int) (*Server, error) {
  return &Server{addr, port, "tcp", 30 * time.Second, 100, nil}, nil
}

func NewTLSServer(addr string, port int, tls *tls.Config) (*Server, error) {
  return &Server{addr, port, "tcp", 30 * time.Second, 100, tls}, nil
}

func NewServerWithTimeout(addr string, port int, timeout time.Duration) (*Server, error) {
  return &Server{addr, port, "tcp", timeout, 100, nil}, nil
}

func NewTLSServerWithMaxConnAndTimeout(addr string, port int, maxconns int, timeout time.Duration, tls *tls.Config) (*Server, error) {
  return &Server{addr, port, "tcp", 30 * time.Second, maxconns, tls}, nil
}

```

我们需要解决这个需要多个 New 这个 Server 对象的函数的问题。

### 解决历程

(1) 将非必填的打包：

```
type Config struct {
    Protocol string
    Timeout  time.Duration
    Maxconns int
    TLS      *tls.Config
}

// Server结构体变成这样
type Server struct {
    Addr string
    Port int
    Conf *Config
}

// New函数变成这样
func NewServer(addr string, port int, conf *Config) (*Server, error) {
    //...
}

//Using the default configuratrion
srv1, _ := NewServer("localhost", 9000, nil)

conf := ServerConfig{Protocol:"tcp", Timeout: 60*time.Duration}
srv2, _ := NewServer("locahost", 9000, &conf)


```

但这种情况需要在 New 中判断 Config 是否为空，有更好的办法。

(2) JAVA 风格的 Builder 模式

```
//使用一个builder类来做包装
type ServerBuilder struct {
  Server
}

func (sb *ServerBuilder) Create(addr string, port int) *ServerBuilder {
  sb.Server.Addr = addr
  sb.Server.Port = port
  //其它代码设置其它成员的默认值
  return sb
}

func (sb *ServerBuilder) WithProtocol(protocol string) *ServerBuilder {
  sb.Server.Protocol = protocol
  return sb
}

func (sb *ServerBuilder) WithMaxConn( maxconn int) *ServerBuilder {
  sb.Server.MaxConns = maxconn
  return sb
}

func (sb *ServerBuilder) WithTimeOut( timeout time.Duration) *ServerBuilder {
  sb.Server.Timeout = timeout
  return sb
}

func (sb *ServerBuilder) WithTLS( tls *tls.Config) *ServerBuilder {
  sb.Server.TLS = tls
  return sb
}

func (sb *ServerBuilder) Build() (Server) {
  return  sb.Server
}

// 用法
sb := ServerBuilder{}
server, err := sb.Create("127.0.0.1", 8080).
  WithProtocol("udp").
  WithMaxConn(1024).
  WithTimeOut(30*time.Second).
  Build()


```

这样其实也不错了，唯一的缺点是 Builder 类似乎有点多余。

(3) Functional Options

```
type Option func(*Server)

func Protocol(p string) Option {
    return func(s *Server) {
        s.Protocol = p
    }
}
func Timeout(timeout time.Duration) Option {
    return func(s *Server) {
        s.Timeout = timeout
    }
}
func MaxConns(maxconns int) Option {
    return func(s *Server) {
        s.MaxConns = maxconns
    }
}
func TLS(tls *tls.Config) Option {
    return func(s *Server) {
        s.TLS = tls
    }
}

// 用法
// 通过可变参数数量解决了需要多个函数签名的问题
func NewServer(addr string, port int, options ...func(*Server)) (*Server, error) {
  srv := Server{
    Addr:     addr,
    Port:     port,
    Protocol: "tcp",
    Timeout:  30 * time.Second,
    MaxConns: 1000,
    TLS:      nil,
  }
  for _, option := range options {
    option(&srv)
  }
  //...
  return &srv, nil
}

s1, _ := NewServer("localhost", 1024)
s2, _ := NewServer("localhost", 2048, Protocol("udp"))
s3, _ := NewServer("0.0.0.0", 8080, Timeout(300*time.Second), MaxConns(1000))


```

这组代码传入一个参数，然后返回一个函数，返回的这个函数会设置自己的 Server 参数。例如，当我们调用其中的一个函数 MaxConns(30) 时，其返回值是一个 func(s\* Server) { s.MaxConns = 30 } 的函数。这个叫高阶函数。在数学上，这有点像是等式变换：f(x) = g(y)。

trpc-go 的各种官方库中，就大量使用了这种方法,例如：

```
opts := []client.Option{
		// 命名空间，不填写默认使用本服务所在环境 namespace
		client.WithNamespace("Production"),
		// 服务名
		//client.WithServiceName("trpc.mpgo.pack_filter.Process"),
		client.WithProtocol("http"),
		client.WithSerializationType(codec.SerializationTypeJSON),
	}

// Option 调用参数工具函数
type Option func(*Options)

// WithNamespace 设置 namespace 后端服务环境 正式环境 Production 测试环境 Development
func WithNamespace(namespace string) Option {
	return func(o *Options) {
		o.SelectOptions = append(o.SelectOptions, selector.WithNamespace(namespace))
	}
}

// 创建ClientProxy, 设置协议为HTTP协议,序列化为Json
httpCli := trpchttp.NewClientProxy("trpc.mpgo.pack_filter.Process", opts...)


// NewClientProxy 新建一个http后端请求代理 必传参数 http服务名: trpc.http.xxx.xxx
// name 后端http服务的服务名，主要用于配置key，监控上报，自己随便定义，格式是: trpc.app.server.service
var NewClientProxy = func(name string, opts ...client.Option) Client {
	c := &httpCli{
		ServiceName: name,
		Client:      client.DefaultClient,
	}
	c.opts = make([]client.Option, 0, len(opts)+1)
	c.opts = append(c.opts, client.WithProtocol("http"))
	c.opts = append(c.opts, opts...)
	return c
}

// httpCli 后端请求结构体
type httpCli struct {
	ServiceName string
	Client      client.Client
	opts        []client.Option
}

```

## 3.控制反转

控制反转（Inversion of Control，loC）是一种软件设计的方法，它的主要思想是把控制逻辑与业务逻辑分开，不要在业务逻辑里写控制逻辑，因为这样会让控制逻辑依赖于业务逻辑，而是反过来，让业务逻辑依赖控制逻辑。
一个开关和电灯的例子：开关就是控制逻辑，电器是业务逻辑。我们不要在电器中实现开关，而是要把开关抽象成一种协议，让电器都依赖它。这样的编程方式可以有效降低程序复杂度，并提升代码重用度。

这个结合实例还是理解得不是特别好。

## 4.修饰器模式

### 解决什么问题

可以很轻松地把一些函数装配到另外一些函数上，让你的代码更加简单，也可以让一些“小功能型”的代码复用性更高。

### 具体内容

简单示例：

```
package main

import "fmt"

func decorator(f func(s string)) func(s string) {

    return func(s string) {
        fmt.Println("Started")
        f(s)
        fmt.Println("Done")
    }
}

func Hello(s string) {
    fmt.Println(s)
}

func main() {
    decorator(Hello)("Hello, World!")
    // 为了易读，下面这么写也可以
    // hello := decorator(Hello)
    // hello("Hello")

}

```

复杂一点的例子：

```
package main

import (
    "fmt"
    "log"
    "net/http"
    "strings"
)

func WithServerHeader(h http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        log.Println("--->WithServerHeader()")
        w.Header().Set("Server", "HelloServer v0.0.1")
        h(w, r)
    }
}

func WithAuthCookie(h http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        log.Println("--->WithAuthCookie()")
        cookie := &http.Cookie{Name: "Auth", Value: "Pass", Path: "/"}
        http.SetCookie(w, cookie)
        h(w, r)
    }
}

func WithBasicAuth(h http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        log.Println("--->WithBasicAuth()")
        cookie, err := r.Cookie("Auth")
        if err != nil || cookie.Value != "Pass" {
            w.WriteHeader(http.StatusForbidden)
            return
        }
        h(w, r)
    }
}

func WithDebugLog(h http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        log.Println("--->WithDebugLog")
        r.ParseForm()
        log.Println(r.Form)
        log.Println("path", r.URL.Path)
        log.Println("scheme", r.URL.Scheme)
        log.Println(r.Form["url_long"])
        for k, v := range r.Form {
            log.Println("key:", k)
            log.Println("val:", strings.Join(v, ""))
        }
        h(w, r)
    }
}
func hello(w http.ResponseWriter, r *http.Request) {
    log.Printf("Recieved Request %s from %s\n", r.URL.Path, r.RemoteAddr)
    fmt.Fprintf(w, "Hello, World! "+r.URL.Path)
}

// 多个修饰器的pipeline
type HttpHandlerDecorator func(http.HandlerFunc) http.HandlerFunc

func Handler(h http.HandlerFunc, decors ...HttpHandlerDecorator) http.HandlerFunc {
    for i := range decors {
        d := decors[len(decors)-1-i] // iterate in reverse
        h = d(h)
    }
    return h
}

func main() {
    // 单个修饰器的简单用法
    http.HandleFunc("/v1/hello", WithServerHeader(WithAuthCookie(hello)))
    // pipeline用法
    http.HandleFunc("/v4/hello", Handler(hello,
                WithServerHeader, WithBasicAuth, WithDebugLog))
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}


```

这个用法有点像 trpc-go 的 filter，可以对比一下。

## 5.pipeline

trpc-go filter 的写法


