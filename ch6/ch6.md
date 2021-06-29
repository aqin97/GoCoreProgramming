# 第六章 反射

反射就是程序能够在运行时动态的查看自己的状态，并允许修改自身的行为。本章主要介绍reflect库的用法，对反射的内部实现只是简单的介绍。GO的反射建立在GO类型系统之上，和接口有紧密的联系，学反射之前要先了解接口。

## 6.1 基本概念

GO的反射基础是接口和类型系统，反射巧妙的借用了实例转换为接口所使用的数据结构，将实例传给内部的空接口，实际上是把实例类型转换为一个接口类型可以表达的数据结构`eface`，反射再根据这个数据结构来访问和操作实例的值和类型。

### 6.1.1 基本数据结构和入口函数

**reflect.Type**接口

这是一个描述类型公共信息的结构rtype

```go
type rtype struct {
 size       uintptr
 ptrdata    uintptr // number of bytes in the type that can contain pointers
 hash       uint32  // hash of type; avoids computation in hash tables
 tflag      tflag   // extra type information flags
 align      uint8   // alignment of variable with this type
 fieldAlign uint8   // alignment of struct field with this type
 kind       uint8   // enumeration for C
 // function for comparing objects of this type
 // (ptr to object A, ptr to object B) -> ==?
 equal     func(unsafe.Pointer, unsafe.Pointer) bool
 gcdata    *byte   // garbage collection data
 str       nameOff // string form
 ptrToThis typeOff // type for pointer to this type, may be zero
}
```

这个结构和runtime的`_type`实际上是一个东西，只是因为包的隔离性质分开定义而已。这个结构实现了reflect.Type接口，GO可以通过reflect.TypeOf()函数来返回一个Type类型的接口，通过这个接口来获取对象的信息。那么为什么返回一个接口而不是rtype实例呢？首先，类型信息是一个只读的信息，不应该动态的修改类型的信息，太不安全了；其次不同的类型，类型定义也不一样，用这个接口进行一个统一的抽象。

```go
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

```

对于reflect.TypeOf(a),传进去的实参a有两种类型，一种是具体类型变量（实例），一种是接口。a是实例的话，返回具体的类型信息；a是接口的话，又分为两种情况：a绑定了具体类型变量，则返回接口a动态类型信息；a没有绑定接口类型的话，则返回接口的静态类型信息。

```go
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

```

reflect.Value用来表示实例的值信息，是一个struct，且提供了一些method给使用者。基础数据结构如下

```go
type Value struct {
    //值的类型
    typ *rtype
    //指向值的指针
    ptr unsafe.Pointer
    flag //标记位，可以用来判断是否可以寻址等
}
```

### 6.1.2 基础类型

Type接口有一个Kind()方法，这个方法返回一个整型枚举值，不同的值代表不同的类型，总共有26个。那么什么是基础类型？举例：`[]int`和`[]string`都是slice类型，即他们的基础类型就是slice。

底层类型和基础类型的区别：基础类型是抽象的类型划分，底层类型是对具体的类型定义的，比如不同的struct类型的基础类型都是struct，但是他们的底层类型可能不同

## 6.2 反射的规则

反射对象Value，Type和类型实例之间相互转换

### 6.2.1 反射api

略

### 6.2.2 反射三定律

- 可以从接口值得到反射对象
- 可以从反射对象得到接口值
- 若修改一个反射对象，则其值必须可以修改

## 6.3 inject库