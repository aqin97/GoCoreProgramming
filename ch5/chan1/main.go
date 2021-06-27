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
