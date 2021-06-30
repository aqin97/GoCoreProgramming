package main

import (
	"fmt"

	"github.com/codegangsta/inject"
)

type S1 interface{}
type S2 interface{}

type Staff struct {
	Name    string `inject`
	Company S1     `inject`
	Level   S2     `inject`
	Age     int    `inject`
}

func main() {
	//创建被注入的实例
	s := Staff{}

	inj := inject.New()
	//实参注入
	inj.Map("tom")
	inj.MapTo("tencent", (*S1)(nil))
	inj.MapTo("t3", (*S2)(nil))
	inj.Map(23)

	//实现对struct注入
	inj.Apply(&s)

	fmt.Println(s)
}
