package main

import (
	"fmt"
	"reflect"
)

type INT int

type A struct {
	a int
}

type B struct {
	b string
}

type Ita interface {
	String() string
}

func (b B) String() string {
	return b.b
}

func main() {
	var a INT = 12
	var b int = 14

	ta := reflect.TypeOf(a)
	tb := reflect.TypeOf(b)

	if ta == tb {
		fmt.Println("ta == tb")
	} else {
		fmt.Println("ta != tb")
	}

	//类型名
	fmt.Println(ta.Name())
	fmt.Println(tb.Name())
	//底层基础类型
	fmt.Println(ta.Kind().String())
	fmt.Println()

	s1 := A{1}
	s2 := B{"hello"}

	fmt.Println(reflect.TypeOf(s1).Name())
	fmt.Println(reflect.TypeOf(s2).Name())

	fmt.Println(reflect.TypeOf(s1).Kind().String())
	fmt.Println(reflect.TypeOf(s2).Kind().String())

	ita := new(Ita)
	var itb Ita = s2

	fmt.Println(reflect.TypeOf(ita).Elem().Name())
	fmt.Println(reflect.TypeOf(ita).Elem().Kind().String())

	fmt.Println(reflect.TypeOf(itb).Name())
	fmt.Println(reflect.TypeOf(itb).Kind().String())
}
