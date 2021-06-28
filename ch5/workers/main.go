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
