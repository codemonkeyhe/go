/*
https://blog.golang.org/defer-panic-and-recover

*/

package main

import "fmt"

func main() {
	fmt.Println("hello world!")
	defer func() { // 必须要先声明defer，否则不能捕获到panic异常
		fmt.Println("before recover")
		if err := recover(); err != nil {
			fmt.Println(err) // 这里的err其实就是panic传入的内容，55
		}
		fmt.Println("after recover")
	}()
	f()

}

func f() {
	fmt.Println("here")
	panic("PANIC ERR")
	fmt.Println("this line is never print")
}
