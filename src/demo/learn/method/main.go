// method project main.go
package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
)

type Vertex struct {
	X, Y float64
}

//Go 没有类。然而，仍然可以在结构体类型上定义方法。
//方法接收者 出现在 func 关键字和方法名之间的参数中。
//在 *Vertex 指针类型 提供abs方法
func (v *Vertex) Abs() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

//你可以对包中的 任意 类型定义任意方法，而不仅仅是针对结构体。
//但是，不能对来自其他包的类型或基础类型定义方法。
type MyFloat float64

//MyFloat 值类型 提供abs方法
func (f MyFloat) Abs() float64 {
	if f < 0 {
		return float64(-f)
	}
	return float64(f)
}

//接口类型是由一组方法定义的集合。
//接口类型的值可以存放实现这些方法的任何值
type Abser interface {
	Abs() float64
}

type Reader interface {
	Read(b []byte) (n int, err error)
}

type Writer interface {
	Write(b []byte) (n int, err error)
}

type ReadWriter interface {
	Reader
	Writer
}

type Person struct {
	Name string
	Age  int
}

//一个普遍存在的接口是 fmt 包中定义的 Stringer。
//Stringer 是一个可以用字符串描述自己的类型。`fmt`包 （还有许多其他包）使用这个来进行输出。
func (p Person) String() string {
	return fmt.Sprintf("%v (%v years)", p.Name, p.Age)
}

//程序使用 error 值来表示错误状态。
//与 fmt.Stringer 类似， error 类型是一个内建接口,该接口声明了Error方法
//与 fmt.Stringer 类似，fmt 包在输出时也会试图匹配 error。
type MyError struct {
	When time.Time
	What string
}

func (e *MyError) Error() string {
	return fmt.Sprintf("at %v, %s",
		e.When, e.What)
}

func run() error {
	return &MyError{
		time.Now(),
		"it didn't work",
	}
}

type Hello struct{}

func (h Hello) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request) {
	fmt.Fprint(w, "Hello!")
}

type Struct struct {
	Greeting string
	Punct    string
	Who      string
}

func (h *Struct) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request) {
	fmt.Fprint(w, "CCCCCC")
}

type String string

func (s String) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request) {
	fmt.Fprint(w, "ssss")
}

func main() {
	v := &Vertex{3, 4}
	fmt.Println(v.Abs())

	f := MyFloat(-math.Sqrt2)
	fmt.Println(f.Abs())

	{
		//定义了一个接口类型的对象a  该接口包含了一个名为Abs,返回float64的函数的声明
		var a Abser
		f := MyFloat(-math.Sqrt2)
		v := Vertex{3, 4}

		a = f  // a MyFloat 实现了 Abser
		a = &v // a *Vertex 实现了 Abser，同时覆盖了上面的实现

		// 下面一行，v 是一个 Vertex（而不是 *Vertex）
		// 所以没有实现 Abser。
		//a = v

		//调用的是  (*Vertex).Abs() MyFloat的实现被覆盖了
		fmt.Println(a.Abs())
	}

	{
		//隐式接口解藕了实现接口的包和定义接口的包：互不依赖。
		var w Writer
		// os.Stdout 实现了 Writer
		w = os.Stdout
		fmt.Fprintf(w, "hello, writer\n")
	}

	//重定义结构体的输出
	a := Person{"Arthur Dent", 42}
	z := Person{"Zaphod Beeblebrox", 9001}
	fmt.Println(a, z)

	if err := run(); err != nil {
	}

	{
		//io.Reader 接口有一个 Read 方法：
		//func (T) Read(b []byte) (n int, err error)
		r := strings.NewReader("Hello, Reader!")
		b := make([]byte, 8)
		for {
			//Read 用数据填充指定的字节 slice，并且返回填充的字节数和错误信息。 在遇到数据流结尾时，返回 io.EOF 错误
			n, err := r.Read(b)
			fmt.Printf("n = %v err = %v b = %v\n", n, err, b)
			fmt.Printf("b[:n] = %q\n", b[:n])
			if err == io.EOF {
				break
			}
		}
	}

	//	var h Hello
	//	err := http.ListenAndServe("localhost:4000", h)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//http://localhost:4001/struct

	http.Handle("/struct", &Struct{"Hello", ":", "Gophers!"})
	http.Handle("/string", String("I'm a frayed knot."))
	log.Fatal(http.ListenAndServe("localhost:4001", nil))

}

//如果去掉*，用值类型的话，则对v的修改不会生效
func (v *Vertex) Scale(f float64) {
	v.X = v.X * f
	v.Y = v.Y * f
}
