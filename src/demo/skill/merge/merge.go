package main

import (
	"fmt"
	"log"
	"sync"
	//"math/rand"
	//"time"
)

/*

https://github.com/campoy/justforfunc/blob/master/26-nil-chans/main.go
https://lingchao.xin/post/why-are-there-nil-channels-in-go.html
*/

func asChan(vs ...int) <-chan int {
	c := make(chan int)
	go func() {
		for _, v := range vs {
			c <- v
			//time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		}
		close(c)
	}()
	return c
}

// 一旦它打印出从 1 到 8 的值, 它将开始永远打印 0.
//如果我们从一个关闭的 channel 接收会发生什么 ? 我们会得到 channel 类型的默认值. 在我们的例子中, 类型是 int, 所以值是 0.
//a和b关闭后，select v:= 0，而且是不停地产生0
func WrongMerge(a, b <-chan int) <-chan int {
	c := make(chan int)
	go func() {
		for {
			select {
			case v := <-a:
				c <- v
			case v := <-b:
				c <- v
			}
		}
	}()
	return c
}

// 似乎一旦一个 channel 完成, 我们就不停地迭代 !
// 毕竟它确实有意义. 正如我们在开始时看到的, 从一个关闭的 channel 读取从不阻塞.
// 因此, 只要两个 channels 都处于打开状态, select 语句将会阻塞, 直到新元素准备就绪,
//  但是一旦其中一个关闭, 我们将迭代并浪费 CPU. 这也被称为繁忙的循环, 并不好.
func BadMerge(a, b <-chan int) <-chan int {
	c := make(chan int)
	go func() {
		//defer 语句位于新的 goroutine 中调用的匿名函数中, 而不是在 merge 中. 否则, 只要我们退出 merage, c 就会被关闭, 那么发送一个值给它将引发 panic.
		//理论上chan应该由发送方关闭
		defer close(c)
		adone, bdone := false, false
		for !adone || !bdone {
			select {
			case v, ok := <-a:
				if !ok {
					// 假设a先遍历完，a的chan已经关闭了，adone已经为true了，但是b的还没关闭，依然有概率不停地打印aIsDone。
					// 主要原因是，当a关闭后，依然有Select在<-a上等待，这个写法本身是错误的
					// 可以参照Pipeline的做法，维护一个全局的done，上游关闭了通道后，通过done传递给下游的消费者，让消费者执行select <-done后的ruturn结束消费行为。
					log.Println("aIsDone")
					adone = true
					continue
				}
				c <- v
			case v, ok := <-b:
				if !ok {
					log.Println("bIsDone")
					bdone = true
					continue
				}
				c <- v
			}
		}
	}()
	return c
}

func merge(a, b <-chan int) <-chan int {
	c := make(chan int)
	go func() {
		defer close(c)
		for a != nil || b != nil {
			select {
			case v, ok := <-a:
				if !ok {
					fmt.Println("aIsDone")
					a = nil
					continue
				}
				c <- v
			case v, ok := <-b:
				if !ok {
					fmt.Println("bIsDone")
					b = nil
					continue
				}
				c <- v
			}
		}
	}()
	return c
}

//借鉴pipeline 不要使用select在多个chan上等待，否则其中一个chan关闭后就会出现前面的问题。
//采用for range在chan上等待，主要是因为for range能自动感知chan的关闭而结束for
//这就是for range和select的区别
func mergeBeter(a, b <-chan int) <-chan int {
	c := make(chan int)
	var wg sync.WaitGroup
	wg.Add(2)
	output := func(in <-chan int) {
		for n := range in { //in的消费者
			//select { //out的新生产者
			//case c <- n:
			//case <-done:
			//}
			c <- n
		}
		wg.Done()
	}
	go output(a)
	go output(b)
	go func() {
		wg.Wait()
		close(c)
	}()
	return c
}

func arrayMerge(left []int, right []int) {
	if len(left) == 0 || len(right) == 0 {
		return
	}
	a := asChan(left...)
	b := asChan(right...)
	c := merge(a, b)
	for v := range c {
		fmt.Print(v, " ")
	}
}

func arrayMergeNormalCopy(left []int, right []int) []int {
	res := make([]int, len(left)+len(right), len(left)+len(right))
	l := copy(res, left)
	//r := copy(res[l:], right)
	copy(res[l:], right)
	//fmt.Println("L=", l, " R=", r)
	//copy(res, right)
	return res
}

func arrayMergeNormal(left []int, right []int) []int {
	res := make([]int, 0, len(left)+len(right))
	for _, v := range left {
		res = append(res, v)
	}
	for _, v := range right {
		res = append(res, v)
	}
	return res
}

//4次内存分配，其中1次chan+1次res,另外2次是？
func arrayMergeBetter(left []int, right []int) []int {
	if len(left) == 0 || len(right) == 0 {
		return nil
	}
	c := make(chan int)
	var wg sync.WaitGroup
	wg.Add(2)
	output := func(arr []int) {
		for _, n := range arr { //in的消费者
			c <- n
		}
		wg.Done()
	}
	go output(left)
	go output(right)
	go func() {
		wg.Wait()
		close(c)
	}()
	res := make([]int, 0, len(left)+len(right))
	for v := range c {
		res = append(res, v)
	}
	return res
}

func arrayMergeBetterV2(left []int, right []int) []int {
	if len(left) == 0 || len(right) == 0 {
		return nil
	}
	done := make(chan struct{})
	c := make(chan int)
	output := func(arr []int, done chan<- struct{}) {
		for _, n := range arr { //in的消费者
			c <- n
		}
		done <- struct{}{}
	}
	go output(left, done)
	go output(right, done)

	go func() {
		<-done
		<-done
		close(c)
	}()

	res := make([]int, 0, len(left)+len(right))
	for v := range c {
		res = append(res, v)
	}
	return res
}

func arrayMergeBetterN(arrs ...[]int) []int {
	if len(arrs) == 0 {
		return nil
	}
	c := make(chan int)
	var wg sync.WaitGroup
	wg.Add(len(arrs))
	output := func(arr []int) {
		for _, n := range arr { //in的消费者
			c <- n
		}
		wg.Done()
	}
	maxsize := 0
	for _, arr := range arrs {
		if len(arr) > maxsize {
			maxsize = len(arr)
		}
		go output(arr)
	}
	go func() {
		wg.Wait()
		close(c)
	}()
	res := make([]int, 0, len(arrs)*maxsize)
	for v := range c {
		res = append(res, v)
	}
	return res
}

func Min(a ...int) int {
	if len(a) == 0 {
		return 0
	}
	min := a[0]
	for _, v := range a {
		if v < min {
			min = v
		}
	}
	return min
}

func Wmin(arr []int) int {
	res := Min(arr...)
	return res
}

func arrayMergeByCopy3(a1 []uint32, a2 []uint32, a3 []uint32) []uint32 {
	size := len(a1) + len(a2) + len(a3)
	res := make([]uint32, size, size)
	l1 := copy(res, a1)
	fmt.Println("L1=", l1)
	l2 := copy(res[l1:], a2)
	fmt.Println("L2=", l2)
	l2 += l1
	fmt.Println("L2=", l2)
	copy(res[l2:], a3)
	return res
}

func inArrays(appid uint32, arrs ...[]uint32) bool {
	for _, arr := range arrs {
		for _, v := range arr {
			if v == appid {
				return true
			}
		}
	}
	return false
}

func main() {
	if true {
		a1 := []uint32{1, 3, 5, 7}
		a2 := []uint32{2, 4, 6, 8, 0, 1}
		a3 := []uint32{11, 45, 33, 78}
		//fmt.Println(arrayMergeByCopy3(a1, a2, a3))
		fmt.Println(inArrays(73, a1, a2, a3))
	}

	if false { //变成参数传参
		arr := []int{7, 9, 3, 5, 1}
		//x := Min(arr...)
		x := Wmin(arr)
		fmt.Printf("The minimum in the array arr is: %d\n", x)
	}

	if false { // merge chan版本
		a := asChan(1, 3, 4, 5, 7)
		b := asChan(2, 4, 6, 8)
		//c := WrongMerge(a, b)
		//c := BadMerge(a, b)
		//c := merge(a, b)
		c := mergeBeter(a, b)
		for v := range c {
			fmt.Print(v, " ")
		}
	}

	if false { // merge slice版本
		left := []int{1, 3, 5, 7}
		right := []int{2, 4, 6, 8, 0, 1}
		//arrayMerge(left, right)
		//fmt.Println(arrayMergeBetter(left, right))
		//fmt.Println(arrayMergeNormal(left, right))
		fmt.Println(arrayMergeNormalCopy(left, right))
		fmt.Println(arrayMergeBetterV2(left, right))
	}

	if false {
		a1 := []int{1, 3, 5, 7}
		a2 := []int{2, 4, 6, 8, 0, 1}
		a3 := []int{11, 45, 33, 78}
		fmt.Println(arrayMergeBetterN(a1, a2, a3))
	}

}
