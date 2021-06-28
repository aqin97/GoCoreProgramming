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
