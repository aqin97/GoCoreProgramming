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
