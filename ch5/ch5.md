# 第5章 并发

## 5.1 并发基础

### 5.1.1 并发和并行

- 并行意味着`在任意时刻`都是同时运行的
- 并发意味着`单位时间内`都是同时运行的

### 5.1.2 goroutine

- 关键字go+匿名函数启动goroutine

```go
package main

import (
 "runtime"
 "time"
)

func main() {
 go func() {
  sum := 0
  for i := 0; i < 10000; i++ {
   sum += i
  }
  println(sum)
  time.Sleep(1 * time.Second)
 }()

 println("numGoroutine=", runtime.NumGoroutine())
 time.Sleep(5 * time.Second)
}

```

- 关键字go+有名函数启动goroutine

```go
package main

import (
 "runtime"
 "time"
)

func sum() {
 sum := 0
 for i := 0; i < 10000; i++ {
  sum += i
 }
 println(sum)
 time.Sleep(2 * time.Second)
}

func main() {
 go sum()

 println("numGoroutine=", runtime.NumGoroutine())
 time.Sleep(5 * time.Second)
}

```

goroutine有这些特性

- 非阻塞的，不会等待（主协程结束会导致其他协程也结束）
- go关键字后面的函数返回值会被忽略（无法通过返回值和其他协程进行通信，需要通过chan）
- 调度器不能保证goroutine的执行顺序（同样的，可以通过chan）
- 没有父子goroutine的说法，所有goroutine都是平等的被调度和执行的
- 程序执行时会为main函数单独创建一个goroutine，遇到其他的go关键字再去创建新的goroutine
- goroutine不暴露id给程序员，不能在一个goroutine中显式的调用另一个goroutine

### 5.1.3 chan

通道是goroutine之间通信和同步的重要组件。GO语言的哲学之一就是“不通过内存共享来通信，而通过通信来共享内存”。声明一个简单的通道语句是`chan dataType`，但是简单的声明是没有意义的，这个通道没有初始化，其值为nil。
通道分为无缓冲的和有缓冲的，go提供len和cap函数，无缓冲管道的len和cap都是0，有缓冲的len表示chan中还有几个元素，cap表示chan的容量上限。无缓冲的通道可以用来通信，也可以用于同步两个goroutine，有缓冲的可以用来通信。
对之前程序进行修改，利用chan而不是time.Sleep()来实现goroutine之间的同步

```go
package main

import (
 "runtime"
)

func main() {
 c := make(chan int)
 go func(c chan int) {
  sum := 0
  for i := 0; i < 10000; i++ {
   sum += i
  }
  println(sum)
  c <- 0
 }(c)

 println("numGoroutine=", runtime.NumGoroutine())

 <-c
}

```

goroutine运行结束后，写到通道里的数据不会消失，所以通道可以缓冲和适配两个goroutine处理速率不同的情况，缓冲通道和消息队列类似，都有削峰和增大吞吐量的功能。
如下：

```go
package main

import (
 "runtime"
 "time"
)

func main() {
 c := make(chan struct{})
 ci := make(chan int, 100)
 go func(c chan struct{}, ci chan int) {
  for i := 0; i < 10; i++ {
   ci <- i
  }
  close(ci)
  c <- struct{}{}
 }(c, ci)
 println("numGoroutine=", runtime.NumGoroutine())
 <-c
 time.Sleep(1 * time.Second)
 println("numGoroutine=", runtime.NumGoroutine())

 for v := range ci {
  println(v)
 }
}

```

操作不同的chan会导致三种情况

#### panic

（1）向已经关闭的chan写数据会导致panic
（2）重复关闭chan也会panic

#### 阻塞

（1）向未初始化的chan写数据或者读数据都会导致当前goroutine永久阻塞
（2）向缓冲区已满的chan中写数据会阻塞
（3）从空的chan中读数据也会阻塞

#### 非阻塞

（1）向有缓冲且没有满的chan中读/写数据都不会阻塞
（2）读取已关闭的chan不会引发阻塞，而是返回对应类型的零值(无缓冲是这样，有缓冲会先把缓冲中的数据输出,输出完毕之后再去读就会返回对应零值)

### 5.1.4 WaitGroup

等待组用来确保所有的goroutine都执行完毕之后再进行下一项任务
使用方法也很简单：

- 声明一个等待组
- 使用Add()方法为等待组的计数器设置值
- 每有一个goroutine完成任务，使用Done()方法对计数器进行减一
- 调用Wait()方法，该方法会阻塞至计数器清0

示例如下：

```go
package main

import (
 "net/http"
 "sync"
)

var wg sync.WaitGroup

var urls = []string{
 "http://www.golang.org",
 "http://www.baidu.com",
 "http://www.qq.com",
}

func main() {
 for _, url := range urls {
  wg.Add(1)
  go func(url string) {
   defer wg.Done()
   resp, err := http.Get(url)
   if err != nil {
    panic(err)
   }
   println(resp.Status)
  }(url)
 }
 wg.Wait()
}

```

### 5.1.5 select

select是类UNIX系统提供的一个多路复用系统API，go语言借用多路复用的概念，提供select关键字，用于多路监听多个通道。当监听的通道中没有可读或者可写的，select阻塞；只要有一个是可读或者可写的，select就不会阻塞；若通道同时有多个可读或者可写的状态，select会随机选择一个执行
示例如下：

```go
package main

func main() {
 ch := make(chan int, 1)
 go func(chan int) {
  for {
   select {
   case ch <- 1:
   case ch <- 0:
   }
  }
 }(ch)

 for i := 0; i < 10; i++ {
  println(<-ch)
 }
}

```

### 5.1.6 扇入（fan in）和扇出（fan out）

扇入：多条通道聚合到一条通道中来，GO中最简单的扇入就是使用select聚合多条通道服务
扇出：将一条通道发散到多条通道中处理，GO语言中具体的实现就是使用go关键字启动多个goroutine
扇入就是合，扇出就是分；
生产者效率低时，需要扇入聚合多个生产者来满足消费者，如耗时的加密解密服务；
消费者效率低时，需要扇出技术，如web服务器并发处理请求。

### 5.1.7 通知退出机制（close channel to broadcast）

读取已经关闭的通道不会引起阻塞，也不会导致panic，而是立即返回该通道存储类型的零值。关闭被select监听的某个通道会被select立即感知到这种通知，然后进行相应处理，这就是退出通知机制。之后的context库就是利用这种机制来处理更复杂的通知的，退出通知机制是学习context库的基础
示例：演示退出通知机制，下游的消费者不需要随机数时，显式的通知生产者停止生产

```go
package main

import (
 "fmt"
 "math/rand"
 "runtime"
)

func GenerateIntA(done chan struct{}) chan int {
 ch := make(chan int)
 go func() {
 Lable:
  for {
   select {
   case ch <- rand.Int():
   case <-done:
    break Lable
   }
  }
  close(ch)
 }()
 return ch
}

func main() {
 done := make(chan struct{})
 ch := GenerateIntA(done)
 fmt.Println(<-ch)
 fmt.Println(<-ch)

 close(done)

 fmt.Println(<-ch)
 fmt.Println(<-ch)
 println("numGoroutine=", runtime.NumGoroutine())
}

```

程序的结果可能是：
5577006791947779410
8674665223082153551
6129484611666145821
0
numGoroutine= 1
或者是：
5577006791947779410
8674665223082153551
0
0
numGoroutine= 1
为什么会有两种结果呢？
关闭done通道之后，select监听的两个通道就都不阻塞了，select会随机选择一个执行，所以会出现两个结果

## 5.2 并发范式

本节通过示例来演示GO的并发处理能力，每个示例代表一个范式，这些范式都有典型的特征，在真是程序中稍稍改造就可以使用。

### 5.2.1 生成器

在应用系统编程中，比较常见的时调用一个统一的全局的生成器服务，用于生成全局事务编号、订单号、序列号和随机数等。
带缓冲的生成器，如下：

```go
package main

import (
 "fmt"
 "math/rand"
)

func GenerateIntA() chan int {
 ch := make(chan int, 10)
 go func() {
  for {
   ch <- rand.Int()
  }
 }()

 return ch
}

func main() {
 ch := GenerateIntA()
 fmt.Println(<-ch)
 fmt.Println(<-ch)
}

```

多个goroutine增强型生成器(使用扇入技术),如下：

```go
package main

import (
 "fmt"
 "math/rand"
)

func GenerateIntA() chan int {
 ch := make(chan int, 10)
 go func() {
  for {
   ch <- rand.Int()
  }
 }()

 return ch
}

func GenerateIntB() chan int {
 ch := make(chan int, 10)
 go func() {
  for {
   ch <- rand.Int()
  }
 }()

 return ch
}

func GenerateInt() chan int {
 ch := make(chan int, 20)
 go func() {
  for {
   select {
   case ch <- <-GenerateIntA():
   case ch <- <-GenerateIntB():
   }
  }
 }()

 return ch
}

func main() {
 ch := GenerateInt()
 for i := 0; i < 100; i++ {
  fmt.Println(<-ch)
 }
}

```

有时又希望生成器可以自动退出，可以借助GO通道的退出通知机制（close channel to broadcast），如下：

```go
package main

import (
 "fmt"
 "math/rand"
)

func GenerateIntA(done chan struct{}) chan int {
 ch := make(chan int)
 go func() {
 Lable:
  for {
   select {
   case ch <- rand.Int():
   case <-done:
    break Lable
   }
  }
  close(ch)
 }()

 return ch
}

func main() {
 done := make(chan struct{})
 ch := GenerateIntA(done)

 fmt.Println(<-ch)
 fmt.Println(<-ch)
 close(ch)
 for v := range ch {
  fmt.Println(v)
 }
}

```

一个融合了并发、缓冲、退出通知等多重特性的生成器，如下：

```go
package main

import (
 "fmt"
 "math/rand"
)

func GenerateIntA(done chan struct{}) chan int {
 ch := make(chan int)
 go func() {
 Lable:
  for {
   select {
   case ch <- rand.Int():
   case <-done:
    break Lable
   }
  }
  close(ch)
 }()

 return ch
}

func GenerateIntB(done chan struct{}) chan int {
 ch := make(chan int, 5)
 go func() {
 Lable:
  for {
   select {
   case ch <- rand.Int():
   case <-done:
    break Lable
   }
  }
  close(ch)
 }()

 return ch
}

func GenerateInt(done chan struct{}) chan int {
 ch := make(chan int, 10)
 send := make(chan struct{})
 go func() {
 Lable:
  for {
   select {
   case ch <- <-GenerateIntA(send):
   case ch <- <-GenerateIntB(send):
   case <-done:
    send <- struct{}{}
    send <- struct{}{}
    break Lable
   }
  }
  close(ch)
 }()
 return ch
}

func main() {
 done := make(chan struct{})
 ch := GenerateInt(done)
 for i := 0; i < 10; i++ {
  fmt.Println(<-ch)
 }
 done <- struct{}{}
 fmt.Println("stop generate")
}

```

### 5.2.2 管道（pipeline）

通道实际上可以分为两个方向，读和写，加入一个函数的输入参数和输出参数都是同一个类型的通道，则该函数可以调用自己，最终形成一个调用链。当然，多个具有相同参数类型的函数也能组成一个调用链，这很像UNIX环境的管道，是一个有类型的管道。
展示GO的链式处理能力，如下：

```go
package main

import "fmt"

func chain(in chan int) chan int {
 out := make(chan int)
 go func() {
  for v := range in {
   out <- v + 1
  }
  close(out)
 }()

 return out
}

func main() {
 in := make(chan int)
 go func() {
  for i := 0; i < 10; i++ {
   in <- i
  }
  close(in)
 }()
 out := chain(chain(chain(in)))
 for v := range out {
  fmt.Println(v)
 }
}
```

### 5.2.3 每个请求都分配一个goroutine

这种并发模式相对比较简单，来一个任务或者请求就启动一个goroutine去处理，典型的是GO中的HTTP server服务。
我们拿计算100个自然数的和来举例，将计算任务拆分为多个task，每个task启动一个goroutine进行处理，代码如下：

```go
package main

import (
 "fmt"
 "sync"
)

type task struct {
 begin  int
 end    int
 result chan<- int
}

//执行任务：计算从begin到end的和，执行结果写入result
func (t *task) do() {
 sum := 0
 for i := t.begin; i <= t.end; i++ {
  sum += i
 }
 t.result <- sum
}

func InitTask(taskchan chan<- task, r chan int, p int) {
 qu := p / 10
 mod := p % 10
 high := qu * 10
 for j := 0; j < qu; j++ {
  b := 10*j + 1
  e := 10 * (j + 1)
  tsk := task{
   begin:  b,
   end:    e,
   result: r,
  }
  taskchan <- tsk
 }
 if mod != 0 {
  tsk := task{
   begin:  high + 1,
   end:    p,
   result: r,
  }
  taskchan <- tsk
 }

 close(taskchan)
}

//将完成初始化的任务，每个task分给一个worker goroutine老处理
//等待所有的task完成，关闭结果通道
func DistributeTask(taskchan <-chan task, wait *sync.WaitGroup, result chan int) {
 for task := range taskchan {
  wait.Add(1)
  go ProcessTask(task, wait)
 }
 wait.Wait()
 close(result)
}

func ProcessTask(t task, wait *sync.WaitGroup) {
 t.do()
 wait.Done()
}

//从结果通道读取，汇总结果
func ProcessResult(resultchan chan int) int {
 sum := 0
 for r := range resultchan {
  sum += r
 }
 return sum
}

func main() {
 taskchan := make(chan task, 10)
 result := make(chan int)
 wait := &sync.WaitGroup{}
 //初始化整个任务
 go InitTask(taskchan, result, 100)
 //分配任务
 go DistributeTask(taskchan, wait, result)
 sum := ProcessResult(result)
 fmt.Println("sum=", sum)
}

```

### 5.2.4 固定worker工作池

GO语言中很容易构建固定数量的goroutines作为工作池线程。
程序中除了主要的main goroutine，还有以下几类goroutine：
（1）初始化任务的goroutine
（2）分发任务的goroutine
（3）等待所有worker结束，然后关闭结果通道的goroutine。
main函数复杂拉起上述goroutine，并从结果通道中获取最后的结果。
程序要采取三个通道：
（1）传递任务task的通道
（2）传递结果result的通道
（3）传递任务结束后关闭所有通道信号的通道
代码如下：

```go
package main

import "fmt"

//固定的workers数量
const (
 NUMBER = 10
)

type task struct {
 begin  int
 end    int
 result chan int
}

//单个计算任务
func (t *task) do() {
 sum := 0
 for i := t.begin; i <= t.end; i++ {
  sum += i
 }
 t.result <- sum
}

func InitTask(taskChan chan<- task, r chan int, p int) {
 qu := p / 10
 mod := p % 10
 high := qu * 10
 for i := 0; i < qu; i++ {
  tsk := task{
   begin:  10*i + 1,
   end:    10 * (i + 1),
   result: r,
  }
  taskChan <- tsk
 }
 if mod != 0 {
  tsk := task{
   begin:  high + 1,
   end:    p,
   result: r,
  }
  taskChan <- tsk
 }
 close(taskChan)
}

//发布任务
func DistributeTask(taskChan <-chan task, workers int, done chan struct{}) {
 for i := 0; i < workers; i++ {
  go ProcessTask(taskChan, done)
 }
}

func ProcessTask(taskChan <-chan task, done chan struct{}) {
 for t := range taskChan {
  t.do()
 }
 done <- struct{}{}
}

func CloseResult(done chan struct{}, resultChan chan int, workers int) {
 for i := 0; i < workers; i++ {
  <-done
 }
 close(done)
 close(resultChan)
}

func ProcessResult(result chan int) int {
 sum := 0
 for res := range result {
  sum += res
 }

 return sum
}

func main() {
 workers := NUMBER
 taskChan := make(chan task, 10)
 resultChan := make(chan int, 10)
 done := make(chan struct{}, 10)

 go InitTask(taskChan, resultChan, 100)
 go DistributeTask(taskChan, workers, done)
 go CloseResult(done, resultChan, workers)

 sum := ProcessResult(resultChan)
 fmt.Println(sum)
}

```

### 5.2.5 future模式

经常遇到在一个流程中需要调用多个子调用的情况，这些子调用之间没有依赖，如果串行的调用，则耗时会很长，可以使用future模式。
基本工作原理：
（1）使用chan作为函数参数
（2）启动goroutine调用函数
（3）通过chan传入参数
（4）做其他可以并行处理的实情
（5）通过chan异步获得结果
代码如下：

```go
package main

import (
 "fmt"
 "time"
)

type query struct {
 sql    chan string
 result chan string
}

func execQuery(q query) {
 go func() {
  sql := <-q.sql
  q.result <- "result from " + sql
 }()
}

func main() {
 q := query{
  make(chan string, 1),
  make(chan string, 1),
 }
 go execQuery(q)
 q.sql <- "select * from table"

 time.Sleep(1 * time.Second)

 fmt.Println(<-q.result)
}

```

future模式的最大好处是将同步调用转换为异步调用，实际情况要比上面复杂的多，要考虑错误和异常的处理。

## 5.3 context标准库

GO语言中的goroutine没有父与子的关系,自然不会有所谓子进程退出后的通知机制,所有的goroutine都是平行的被调度，多个goroutine如何协作涉及通信，同步，通知和退出四个方面。
**通信**：chan是通信基础，这里的通信指程序的数据通道
**同步**：无缓冲的chan是一种同步机制，同样的sync.WaitGroup也是一种同步机制
**通知**：这个通知和上面的通信不太一样，通信负责的是业务数据，而通知往往处理管理、控制流数据。一个简单实现，输入端绑定两个chan，一个处理业务，一个处理通知，用select收敛进行处理。这是一种简单的实现，而不是一种通用的解决办法。
**退出**：借助一个单独的通道和select实现close channel to broadcast。
GO语言在遇到复杂的并发结构处理起来就力不从心。实际编程中goroutine会拉起新的goroutine，新的再拉起新的，最终形成一个树状结构，由于goroutine中没有父子关系，这个树状结构只是在程序员脑中抽象出来的，程序的执行模型没有维护这么一个树状结构。如何通知树上所有的goroutine都退出呢?仅靠语法层面是很难实现的。GO推出了context标准库，提供两种功能：退出通知和元数据调用。

### 5.3.1 context设计目的

跟踪goroutine调用树，并在调用树中传递通知和元数据

### 5.3.2 工作机制和基本数据结构

整体工作机制：第一个创建Context的goroutine被称为root节点。root节点负责创建一个实现Context接口的具体对象，并将该对象作为参数传递到新拉起来的goroutine中，下游的goroutine可以继续封装该对象，再传递给更下游。Context对象在传递过程中最终形成一个树状的数据结构，这样位于root节点的Context对象就能遍历整个Context树，通知和消息就能通过root节点传递出去，实现了上游对下游的消息传递。

Context接口

```go
type Context interface {
    //是否实现超时控制，实现则ok返回true，deadline为超时时间
    //否则ok返回false
    Deadline() (deadline time.Time, ok bool)
    //后面被调用的goroutine应该监听该方法返回的chan，以便及时释放资源
    Done() <-chan struct{}
    //done返回的chan收到通知以后，才可以访问Err()获取因什么原因被取消
    Err() error
    //可以访问上游goroutine给下游穿的的goroutine的值
    Value(key interface{}) interface{}
}
```

canceler接口
canceler接口是一个拓展接口，规定了取消通知的Context具体类型需要实现的接口。context包中的具体类型`*cancelCtx`和`*timerCtx`都实现了该接口。示例如下：

```go
type canceler interface {
    //创建cancel接口实例的goroutine调用cancel方法后通知后续创建的goroutine退出
    cancel(removeFromParent bool, err error)
    //Done方法返回的chan要后端的goroutine来监听， 并及时退出
    Done() <-chan struct{}
}
```

emptyContext结构
这个结构实现了Context接口，但不具备任何功能，因为其所有的方法都是空实现。其存在目的是作为Context对象树的根（root节点）。context包的使用思路就是不停的调用context包提供的包装函数来创建具有特殊功能的Context实例，每个Context实例的创建都以上一个Context对象作为参数，最终形成一个树状结构。示例如下：

```go
type emptyCtx int

func (*emptyCtx) Deadline() (deadline time.Time, ok bool) { return }

func (*emptyCtx) Done() <-chan struct{} { return nil }

func (*emptyCtx) Err() error { return nil }

func (*emptyCtx) Value(key interface{}) interface{} { return nil}

func (e *emptyCtx) String() string {
    switch e {
    case background:
        return "context.background"
    case todo:
        return "context,TODO"
    }
}

```

在context包中定义了两个全局变量和两个封装函数，返回两个emptyCtx实例对象，实际使用时用这两个封装函数来构造Context的root节点。示例如下：

```go
var (
    background = new(emptyCtx)
    todo = new(emptyCtx)
)

func Background() Context {
    return background
}

func TODO() Context {
    return todo
}

```

cancelCtx是一个实现了Context接口和cancler接口的具体类型，canceler具有退出通知方法。退出通知机制不仅能通知自己，也能逐层通知其children节点。示例如下：

```go
type cancelCtx struct {
    Context
    done chan struct{}
    mu sync.Mutex
    children map[canceler]bool
    err error
}

func (c *cancelCtx) Done() <-chan struct{} {
    return c.done
}

func (c *cancelCtx) Err() error {
    c.mu.Lock()
    defer c.mu.UnLock()
    return c.err
}

func (c *cancelCtx) String() string {
    return fmt.Sprintf("%v.WithCancel", c.Context)
}

func (c *cancelCtx) cancel(removeFromParent bool, err, error) {
    if err == nil {
        panic("context: internal error: missing cancel error")
    }
    c.mu.Lock()
    if c.err != nil {
        c.mu.UnLock()
        return
    }
    c.err = err
    close(c.done)

    for child := range c.children {
        chiid.cancel(false, err)
    }
    c.children = nil
    c.mu.UnLock()

    if removeFromParent {
        removeChild(c.Context, c)
    }
}

```
