// hello project main.go
package main

import (
	"fmt"
	"math"
	"math/cmplx"
	"math/rand"
	"runtime"
	"strings"
	"time"

	//"golang.org/x/tour/wc"
)

func add(x int, y int) int {
	return x + y
}

func add1(x, y int) int {
	return x + y
}

//函数可以返回任意数量的返回值
func swap(x, y string) (string, string) {
	return y, x
}

//返回值可以被命名，并且就像在函数体开头声明的变量那样使用。
func split(sum int) (x, y int) {
	x = sum - 3
	y = sum - 4
	return
}

var c, python, java bool

func needInt(x int) int { return x*10 + 1 }
func needFloat(x float64) float64 {
	return x * 0.1
}

//if 语句可以在条件之前执行一个简单语句
//由这个语句定义的变量的作用域仅在 if 范围之内
func pow(x, n, lim float64) float64 {
	if v := math.Pow(x, n); v < lim {
		return v
	}
	return lim
}

func printBoard(s [][]string) {
	for i := 0; i < len(s); i++ {
		fmt.Printf("%s\n", strings.Join(s[i], " "))
	}
}

func printSlice(s string, x []int) {
	fmt.Printf("%s len=%d cap=%d %v\n",
		s, len(x), cap(x), x)
}

func WordCount(s string) map[string]int {
	m := make(map[string]int)
	vs := strings.Fields(s)
	for _, v := range vs {
		val, ok := m[v]
		if ok {
			m[v] = val + 1
		} else {
			m[v] = 1
		}
	}
	return m
}

//函数也是值。他们可以像其他值一样传递，比如，函数值可以作为函数的参数或者返回值。
func compute(fn func(float64, float64) float64) float64 {
	return fn(3, 4)
}

func main() {
	fmt.Println("Hello World!")
	fmt.Println("My favorite number is", rand.Intn(10))
	fmt.Println("Now you have %g problems.", math.Sqrt(7))
	fmt.Println(math.Pi)
	fmt.Println(add(42, 13))
	fmt.Println(add1(42, 13))

	a, b := swap("hello", "world")

	fmt.Println(a, b)

	fmt.Println(split(17))

	var i int
	fmt.Println(i, c, python, java)
	{
		//在函数中， := 简洁赋值语句在明确类型的地方，可以用于替代 var 定义
		//:= 结构不能使用在函数外。
		var i, j int = 1, 2
		k := 3
		c, python, java := true, false, "no!"
		fmt.Println(i, j, k, c, python, java)
	}

	{
		// 基本类型
		var (
			ToBe   bool       = false
			MaxInt uint64     = 1<<64 - 1
			z      complex128 = cmplx.Sqrt(-5 + 12i)
		)

		const f = "%T(%v)\n"
		fmt.Printf(f, ToBe, ToBe)
		fmt.Printf(f, MaxInt, MaxInt)
		fmt.Printf(f, z, z)

	}

	{
		//零值是：
		//数值类型为 0 ，
		//布尔类型为 false ，
		//字符串为 "" （空字符串
		var i int
		var f float64
		var b bool
		var s string
		fmt.Printf("%v %v %v %q\n", i, f, b, s)
	}

	{
		//类型转换 表达式 T(v) 将值 v 转换为类型 T
		var x, y int = 3, 4
		var f float64 = math.Sqrt(float64(x*x + y*y))
		var z uint = uint(f)
		fmt.Println(x, y, z)
	}

	{
		//类型推导
		i := 42           // int
		f := 3.142        // float64
		g := 0.867 + 0.5i // complex128
		fmt.Printf("i is of type %T\n", i)
		fmt.Printf("f is of type %T\n", f)
		fmt.Printf("g is of type %T\n", g)
	}

	{
		//常量的定义与变量类似，只不过使用 const 关键字
		//常量不能使用 := 语法定义
		const Pi = 3.14
		const World = "世界"
		fmt.Println("Hello", World)
		fmt.Println("Happy", Pi, "Day")
		const Truth = true
		fmt.Println("Go rules?", Truth)
	}

	{
		//数值常量  高精度的 值 。
		const (
			Big   = 1 << 100
			Small = Big >> 99
		)
		fmt.Println(needInt(Small))
		fmt.Println(needFloat(Small))
		fmt.Println(needFloat(Big))
	}

	//循环体必须用 { } 括起来。
	sum := 0
	for i := 0; i < 10; i++ {
		sum += i
	}
	fmt.Println(sum)
	//for 是 Go 的 “while”
	{
		sum := 1
		for sum < 1000 {
			sum += sum
		}
		fmt.Println(sum)
	}
	//死循环
	//	for {
	//	}
	fmt.Println(
		pow(3, 2, 10),
		pow(3, 3, 20),
	)

	{
		//除非以 fallthrough 语句结束，否则分支会自动终止
		fmt.Print("Go runs on ")
		switch os := runtime.GOOS; os {
		case "darwin":
			fmt.Println("OS X.")
		case "linux":
			fmt.Println("Linux.")
		default:
			// freebsd, openbsd,
			// plan9, windows...
			fmt.Printf("%s.", os)
		}
	}
	fmt.Println()
	{

		today := time.Now().Weekday()
		fmt.Println(today)
		fmt.Println("When's Saturday?")
		switch time.Saturday {
		case today + 0:
			fmt.Println("Today.")
		case today + 1:
			fmt.Println("Tomorrow.")
		case today + 2:
			fmt.Println("In two days.")
		default:
			fmt.Println("Too far away.")
		}
	}

	{
		//没有条件的 switch 同 switch true 一样。
		t := time.Now()
		switch {
		case t.Hour() < 12:
			fmt.Println("Good morning!")
		case t.Hour() < 17:
			fmt.Println("Good afternoon.")
		default:
			fmt.Println("Good evening.")
		}
	}
	{
		//defer 语句会延迟函数的执行直到上层作用域函数返回。  相当于析构函数 提供垃圾回收的机制
		//延迟调用的参数会立刻生成，但是在上层函数返回前函数都不会被调用。
		defer fmt.Println("world")
		fmt.Println("hello")
	}
	fmt.Println()
	fmt.Println()
	func() {
		//延迟的函数调用被压入一个栈中。当函数返回时， 会按照后进先出的顺序调用被延迟的函数调用。
		fmt.Println("counting")
		for i := 0; i < 10; i++ {
			defer fmt.Println(i)
		}
		fmt.Println("done")
	}()

	{
		// C 不同，Go 没有指针运算。
		i, j := 42, 2701
		p := &i         // point to i
		fmt.Println(*p) // read i through the pointer
		*p = 21         // set i through the pointer
		fmt.Println(i)  // see the new value of i

		p = &j         // point to j
		*p = *p / 37   // divide j through the pointer
		fmt.Println(j) // see the new value of j
	}

	//结构体字段
	type Vertex struct {
		X int
		Y int
	}
	{

		fmt.Println(Vertex{1, 2})
		v := Vertex{1, 2}
		v.X = 4
		fmt.Println(v.X)
		//结构体指针  通过指针间接的访问是透明的
		p := &v
		p.X = 1e9
		fmt.Println(v)
	}

	{
		var (
			v1 = Vertex{1, 2}  // 类型为 Vertex
			v2 = Vertex{X: 1}  // Y:0 被省略
			v3 = Vertex{}      // X:0 和 Y:0
			p  = &Vertex{1, 2} // 类型为 *Vertex
		)
		fmt.Println(v1, p, v2, v3)
	}

	{
		//类型 [n]T 是一个有 n 个类型为 T 的值的数组。
		var a [2]string
		a[0] = "Hello"
		a[1] = "World"
		fmt.Println(a[0], a[1])
		fmt.Println(a)
	}

	{
		//一个 slice 会指向一个序列的值，并且包含了长度信息。
		//[]T 是一个元素类型为 T 的 slice。
		//len(s) 返回 slice s 的长度。
		s := []int{2, 3, 5, 7, 11, 13}
		fmt.Println("s ==", s)

		for i := 0; i < len(s); i++ {
			fmt.Printf("s[%d] == %d\n", i, s[i])
		}
	}

	{
		//slice 的 slice
		// Create a tic-tac-toe board.
		game := [][]string{
			[]string{"_", "_", "_"},
			[]string{"_", "_", "_"},
			[]string{"_", "_", "_"},
		}

		// The players take turns.
		game[0][0] = "X"
		game[2][2] = "O"
		game[2][0] = "X"
		game[1][0] = "O"
		game[0][2] = "X"

		printBoard(game)
	}

	{
		s := []int{2, 3, 5, 7, 11, 13}
		fmt.Println("s ==", s)
		fmt.Println("s[1:4] ==", s[1:4])

		// 省略下标代表从 0 开始
		fmt.Println("s[:3] ==", s[:3])

		// 省略上标代表到 len(s) 结束
		fmt.Println("s[4:] ==", s[4:])
	}

	{
		//构造 slice  slice 由函数 make 创建。这会分配一个全是零值的数组并且返回一个 slice 指向这个数组
		a := make([]int, 5)
		printSlice("a", a)
		b := make([]int, 0, 5)
		printSlice("b", b)
		c := b[:2]
		printSlice("c", c)
		d := c[2:5]
		printSlice("d", d)
	}

	{

		//slice 的零值是 nil 。	一个 nil 的 slice 的长度和容量是 0
		var z []int
		fmt.Println(z, len(z), cap(z))
		if z == nil {
			fmt.Println("nil!")
		}
	}

	{
		//for 循环的 range 格式可以对 slice 或者 map 进行迭代循环
		var pow = []int{1, 2, 4, 8, 16, 32, 64, 128}
		//第一个是当前下标（序号）,即i，第二个是该下标所对应元素的一个拷贝,v
		for i, v := range pow {
			fmt.Printf("2**%d = %d\n", i, v)
		}
	}

	{
		//可以通过赋值给 _ 来忽略序号和值
		pow := make([]int, 10)
		for i := range pow {
			pow[i] = 1 << uint(i)
		}
		for _, value := range pow {
			fmt.Printf("%d\n", value)
		}
	}

	{
		type Vertex struct {
			Lat, Long float64
		}
		var m map[string]Vertex

		{
			m = make(map[string]Vertex)
			m["Bell Labs"] = Vertex{
				40.68433, -74.39967,
			}
			fmt.Println(m["Bell Labs"])
		}

		{
			var m = map[string]Vertex{
				"Bell Labs": Vertex{
					40.68433, -74.39967,
				},
				"Google": Vertex{
					37.42202, -122.08408,
				},
			}
			fmt.Println(m)
		}
		{
			var m = map[string]Vertex{
				"Bell Labs": {40.68433, -74.39967},
				"Google":    {37.42202, -122.08408},
			}
			fmt.Println(m)
		}

		{

			m := make(map[string]int)

			m["Answer"] = 42
			fmt.Println("The value:", m["Answer"])

			m["Answer"] = 48
			fmt.Println("The value:", m["Answer"])
			//删除元素
			delete(m, "Answer")
			fmt.Println("The value:", m["Answer"])
			//通过双赋值检测某个键存在 如果 key 在 m 中， ok 为 true。否则， ok 为 false，并且 elem 是 map 的元素类型的零值
			v, ok := m["Answer"]
			fmt.Println("The value:", v, "Present?", ok)
		}
		fmt.Printf("Fields are: %q", strings.Fields("  foo bar  baz   "))
	}
	//wc.Test(WordCount)
	fmt.Println()
	{
		hypot := func(x, y float64) float64 {
			return math.Sqrt(x*x + y*y)
		}
		fmt.Println(hypot(5, 12))

		fmt.Println(compute(hypot))
		fmt.Println(compute(math.Pow))
	}

}
