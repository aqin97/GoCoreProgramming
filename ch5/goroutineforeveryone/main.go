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
