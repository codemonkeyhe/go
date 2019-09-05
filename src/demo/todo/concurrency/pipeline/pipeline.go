/*
Go Concurrency Patterns: Pipelines and cancellation
https://blog.golang.org/pipelines

*/

package main

import (
	"fmt"
	"sync"
	//"time"
)

// http://www.oschina.net/translate/go-concurrency-patterns-pipelines?lang=chs&page=1#

//生产者生成完毕后，立刻关闭了out这个chan，那么消费者在一个已关闭的chan上，也就是out上，执行select <-out的话，将会不停地得到0值，且不会阻塞。参见merge示例。
func gen(nums ...int) <-chan int {
	out := make(chan int)
	//带缓冲区的chan
	// out := make(chan int, len(nums))
	//这样即使消费者堵塞了，也不会阻塞生产者
	go func() {
		for _, n := range nums {
			out <- n
		}
		close(out)
	}()
	return out
}

func genV2(done <-chan struct{}, nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		for _, n := range nums {
			select {
			case out <- n:
			case <-done:
				return
			}
		}
		close(out)
	}()
	return out
}

func sq(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		for n := range in {
			out <- n * n
		}
		close(out)
	}()
	return out
}

func sqV2(done <-chan struct{}, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			select {
			case out <- n * n:
			case <-done:
				return
			}
		}
	}()
	return out
}

func main() {

	if false {
		// Set up the pipeline.
		// 第一阶段：gen，以从列表读出整数的方式转换整数列表到一个通道。gen函数开始goroutine后， 在通道上发送整数并且在在所有的值被发送完后将通道关闭：
		c := gen(2, 3)
		// 第二阶段：sq，从通道接受整数，然后将接受到的每个整数值的平方后返回到一个通道 。在入境通道关闭和发送所有下行值的阶段结束后，关闭出口通道：
		out := sq(c)
		// main函数建立了通道并运行最后一个阶段：它接受来自第二阶段的值并打印出每个值，直到通道关闭：
		// Consume the output.
		fmt.Println(<-out) // 4
		fmt.Println(<-out) // 9
	}

	if false {
		// Set up the pipeline and consume the output.
		// 由于sq有相同类型的入境和出境通道，我们可以写任意次。我们也可以重写main函数，像其他阶段一样做一系列循环
		for n := range sq(sq(gen(2, 3))) {
			fmt.Println(n) // 16 then 81
		}
	}

	if false { //扇出，扇入  相当于map-reduce fan-out即分发任务，即map阶段，fan-in即收集结果，相当于reduce阶段
		/*
			扇出（fan-out）：多个函数能从相同的通道中读数据，直到通道关闭；这提供了一种在一组“人员”中分发任务的方式，使得CPU和I/O的并行处理.
			扇入（fan-in）：一个函数能从多个输入中读取并处理数据，而这多个输入通道映射到一个单通道，该单通道随着所有输入的结束而关闭。
		*/
		in := gen(2, 3)

		// Distribute the sq work across two goroutines that both read from in.
		c1 := sq(in)
		c2 := sq(in)

		// Consume the merged output from c1 and c2.
		// 引入了一个新函数merge去扇入结果
		for n := range merge(c1, c2) {
			fmt.Println(n) // 4 then 9, or 9 then 4
		}
	}

	if false {
		//In our example pipeline, if a stage fails to consume all the inbound values,
		//the goroutines attempting to send those values will block indefinitely:
		in := gen(2, 3)
		c1 := sq(in)
		c2 := sq(in)

		// Consume the first value from output.
		//这里只消费一个值，那么入口处的发送值的地方将堵塞,因为发送不出去，所以发送方也无法关闭管道
		//可以在发送方改为使用带缓冲区的管道，这样发送方不会堵塞，并且发送完就关闭管道。
		//但是如果发送方的数据量大于缓冲区大小，依然会有同样的问题。也就是治标不治本
		out := merge(c1, c2)
		fmt.Println(<-out) // 4 or 9
		return
		// Since we didn't receive the second value from out,
		// one of the output goroutines is hung attempting to send it.

		//This is a resource leak: goroutines consume memory and runtime resources,
		//and heap references in goroutine stacks keep data from being garbage collected.
		//Goroutines are not garbage collected; they must exit on their own.
		//这是一个资源锁：goroutine消耗内存和运行资源，并且在goroutine栈中的堆引用防止数据被回收。Goroutine不能垃圾回收；它们必须自己退出。
	}
	//显式取消
	if false {
		in := gen(2, 3)

		// Distribute the sq work across two goroutines that both read from in.
		c1 := sq(in)
		c2 := sq(in)

		// Consume the first value from output.
		done := make(chan struct{}, 2)
		out := mergeV1(done, c1, c2)
		fmt.Println(<-out) // 4 or 9

		// Tell the remaining senders we're leaving.
		done <- struct{}{}
		done <- struct{}{}
		//例子里有两个受阻的发送方，所以发送的值有两组：
		//因为2,3可能被发往同一个channel,所以就可能会有两个阻塞的output。正常情况下,即2发往c1,3发往c2,只有一个阻塞的output。看消费的是c1还是C2了，没被消费的会堵塞
		//假设2,3都被发送给c1了。那么c1和c2就都阻塞了, C2因为没有数据消费而导致发送方sq堵塞，C1因为有2个，只消费了一个，第二个发送也堵塞了。
		//同时mergeV1里面的分支可能会在C1 c2上同时堵塞，因此要发送2个done
		//done的数组与mergev1的会堵塞的chan数目一致，可认为是参数chan的数目
		//下游的接收者main需要知道潜在会被阻塞的上游发送者sq的数量。追踪这些数量不仅枯燥，还容易出错。也就是得维护发送多少个结束信号到done通道
	}

	if false {
		// Set up a done channel that's shared by the whole pipeline,
		// and close that channel when this pipeline exits, as a signal
		// for all the goroutines we started to exit.
		done := make(chan struct{})
		defer close(done)
		/*
			在GO里面我们通过关闭一个通道来实现，因为一个在已关闭通道上的接收操作总能立即执行，并返回该元素类型的零值。
			这意味着main函数只需关闭“done”通道就能开启所有发送者。close实际上是传给发送者的一个广播信号。
			我们扩展每一个管道函数接收“done”参数并通过一个“defer”语句触发“close”，
			这样所有来自main的返回路径都会以信号通知管道退出。
			done相当于未来的context组件
		*/
		in := genV2(done, 2, 3)

		// Distribute the sq work across two goroutines that both read from in.
		c1 := sqV2(done, in)
		c2 := sqV2(done, in)

		// Consume the first value from output.
		out := mergeV2(done, c1, c2)
		fmt.Println(<-out) // 4 or 9

		// done will be closed by the deferred call.
	}

}

/*
There is a pattern to our pipeline functions:
	stages close their outbound channels when all the send operations are done.
	stages keep receiving values from inbound channels until those channels are closed.

Here are the guidelines for pipeline construction:
	stages close their outbound channels when all the send operations are done.
	stages keep receiving values from inbound channels until those channels are closed or the senders are unblocked.
	Pipelines unblock senders either by ensuring there's enough buffer for all the values that are sent
	or by explicitly signalling senders when the receiver may abandon the channel.

*/
// 缺点: 返回的out如果不被完全消费，则生产者out将会阻塞，因为out是无缓冲的。
// 这种写法：下游必须使用for range把out全部消费光
func merge(cs ...<-chan int) <-chan int {
	var wg sync.WaitGroup
	out := make(chan int)

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan int) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

//mergeV1的done是有缓冲的，缓冲大小为len(cs)。收到done的信号只是为了不阻塞在select上
func mergeV1(done <-chan struct{}, cs ...<-chan int) <-chan int {
	var wg sync.WaitGroup
	out := make(chan int)

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c is closed or it receives a value
	// from done, then output calls wg.Done.
	output := func(c <-chan int) {
		for n := range c { //c的消费者
			select { //out的新生产者
			case out <- n:
			case <-done:
			}
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait() //等待2个chan都消费完毕了，才关闭out
		close(out)
	}()
	return out
}

//mergeV2的done是无缓冲的，相当于广播关闭信号。所以收到done的关闭信号立刻return
func mergeV2(done <-chan struct{}, cs ...<-chan int) <-chan int {
	var wg sync.WaitGroup
	out := make(chan int)

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c is closed or it receives a value
	// from done, then output calls wg.Done.
	output := func(c <-chan int) {
		defer wg.Done()
		for n := range c {
			select {
			case out <- n:
			case <-done:
				return
			}
		}
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
