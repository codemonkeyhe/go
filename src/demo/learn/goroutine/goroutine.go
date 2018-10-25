// goroutine project goroutine.go
package main

import (
	"fmt"
	"sync"
	"time"
)

func say(s string) {
	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Millisecond)
		fmt.Println(s)
	}
}

//channel 是有类型的管道，可以用 channel 操作符 <- 对其发送或者接收值。
func sum(a []int, c chan int) {
	sum := 0
	for _, v := range a {
		sum += v
	}
	c <- sum // 将和送入 c
}

func fibonacci(n int, c chan int) {
	x, y := 0, 1
	for i := 0; i < n; i++ {
		c <- x
		x, y = y, x+y
	}
	close(c)
}

func fibonacci1(c, quit chan int) {
	x, y := 0, 1
	for {
		//select 语句使得一个 goroutine 在多个通讯操作上等待
		//select 会阻塞，直到条件分支中的某个可以继续执行，这时就会执行那个条件分支。当多个都准备好的时候，会随机选择一个。
		select {
		case c <- x:
			x, y = y, x+y
		case <-quit:
			fmt.Println("quit")
			return
		}
	}
}

// SafeCounter 的并发使用是安全的。
type SafeCounter struct {
	v   map[string]int
	mux sync.Mutex
}

// Inc 增加给定 key 的计数器的值。
//保证在每个时刻，只有一个 goroutine 能访问一个共享的变量从而避免冲突
func (c *SafeCounter) Inc(key string) {
	c.mux.Lock()
	// Lock 之后同一时刻只有一个 goroutine 能访问 c.v
	c.v[key]++
	c.mux.Unlock()
}

// Value 返回给定 key 的计数器的当前值。
func (c *SafeCounter) Value(key string) int {
	c.mux.Lock()
	// Lock 之后同一时刻只有一个 goroutine 能访问 c.v
	//defer 语句来保证互斥锁一定会被解锁
	defer c.mux.Unlock()
	return c.v[key]
}

func main() {
	//goroutine 是由 Go 运行时环境管理的轻量级线程。
	//在新的 goroutine 中运行Say
	go say("world")
	say("hello")

	/*
		a := []int{7, 2, 8, -9, 4, 0}
		c := make(chan int)
		go sum(a[:len(a)/2], c)
		go sum(a[len(a)/2:], c)
		x, y := <-c, <-c // 从 c 中获取

		fmt.Println(x, y, x+y)

		//channel 可以是 带缓冲的。为 make 提供第二个参数作为缓冲长度来初始化一个缓冲 channel：
		//向带缓冲的 channel 发送数据的时候，只有在缓冲区满的时候才会阻塞。 而当缓冲区为空的时候接收操作会阻塞
		ch := make(chan int, 2)
		ch <- 1
		ch <- 2
		fmt.Println(<-ch)
		fmt.Println(<-ch)

		{
			c := make(chan int, 10)
			go fibonacci(cap(c), c)
			//循环 `for i := range c` 会不断从 channel 接收值，直到它被关闭
			//channel 与文件不同；通常情况下无需关闭它们。只有在需要告诉接收者没有更多的数据的时候才有必要进行关闭，例如中断一个 range。
			for i := range c {
				fmt.Println(i)
			}
		}

		{
			c := make(chan int)
			quit := make(chan int)
			go func() {
				for i := 0; i < 10; i++ {
					fmt.Println(<-c)
				}
				quit <- 0
			}()
			fibonacci1(c, quit)

		}

		//	{
		//		tick := time.Tick(100 * time.Millisecond)
		//		boom := time.After(500 * time.Millisecond)
		//		for {
		//			//当 select 中的其他条件分支都没有准备好的时候，default 分支会被执行。
		//			//为了非阻塞的发送或者接收，可使用 default 分支：
		//			select {
		//			case <-tick:
		//				fmt.Println("tick.")
		//			case <-boom:
		//				fmt.Println("BOOM!")
		//				//return
		//			default:
		//				fmt.Println("    .")
		//				time.Sleep(50 * time.Millisecond)
		//			}
		//		}
		//	}

		{
			c := SafeCounter{v: make(map[string]int)}
			for i := 0; i < 1000; i++ {
				go c.Inc("somekey")
			}

			time.Sleep(time.Second)
			fmt.Println(c.Value("somekey"))

		}

	*/

}
