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

```
