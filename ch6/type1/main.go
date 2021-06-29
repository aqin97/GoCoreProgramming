package main

import (
	"fmt"
	"reflect"
)

type Student struct {
	Name string `学生姓名`
	Age  int    `a:"1111"b:"3333"`
}

func main() {
	s := Student{}
	rt := reflect.TypeOf(s)
	fieldName, ok := rt.FieldByName("Name")

	//取tag
	if ok {
		fmt.Println(fieldName.Tag)
	}

	fieldAge, ok := rt.FieldByName("Age")

	if ok {
		fieldAge.Tag.Get("a")
		fieldAge.Tag.Get("b")
	}

	fmt.Println("type_name", rt.Name())
	fmt.Println("type_numfield", rt.NumField())
	fmt.Println("type_pgkpath", rt.PkgPath())
	fmt.Println("type_string", rt.String())
	fmt.Println("type_kind_string", rt.Kind().String())

	for i := 0; i < rt.NumField(); i++ {
		fmt.Printf("type.Field[%d].Name = %v\n", i, rt.Field(i).Name)
	}

}
