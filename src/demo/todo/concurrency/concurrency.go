package main

import (
	"fmt"
	"math/rand"
	"time"
)

/*
Concurrency is not parallelism
https://blog.golang.org/concurrency-is-not-parallelism
https://talks.golang.org/2012/concurrency.slide#7
Rob Pike
Google
http://golang.org/s/plusrob

*/

func boring1(msg string) {
	for i := 0; ; i++ {
		fmt.Println(msg, i)
		//time.Sleep(time.Second)
		time.Sleep(time.Duration(rand.Intn(1e3)) * time.Millisecond)
	}
}

func boring2(msg string, c chan string) {
	for i := 0; ; i++ {
		c <- fmt.Sprintf("%s %d", msg, i) // Expression to be sent can be any suitable value.
		time.Sleep(time.Duration(rand.Intn(1e3)) * time.Millisecond)
	}
}

//相当于 pipeline的merge
func fanIn(input1, input2 <-chan string) <-chan string {
	c := make(chan string)
	go func() {
		for {
			c <- <-input1
		}
	}()
	go func() {
		for {
			c <- <-input2
		}
	}()
	return c
}

//引入了select，解释了select优势，对于fanin5来说，减少了协程数目
// Only one goroutine is needed
// 但是对于pipeline的merge，合并N条chan的结果时，select行不通，select只能在多条有名的chan上等待，对于cs ...<-chan int这种chan可变数目的列表来说无能威力
func fanIn5(input1, input2 <-chan string) <-chan string {
	c := make(chan string)
	go func() {
		for {
			select {
			case s := <-input1:
				c <- s
			case s := <-input2:
				c <- s
			}
		}
	}()
	return c
}

// Generator: function that returns a channel
func boring3(msg string) <-chan string { // Returns receive-only channel of strings.
	c := make(chan string)
	go func() { // We launch the goroutine from inside the function.
		for i := 0; ; i++ {
			c <- fmt.Sprintf("%s %d", msg, i)
			time.Sleep(time.Duration(rand.Intn(1e3)) * time.Millisecond)
		}
	}()
	return c // Return the channel to the caller.
}

type Message struct {
	str  string
	wait chan bool
}

// Generator: function that returns a channel
func boring4(msg string, waitForIt chan bool) <-chan Message { // Returns receive-only channel of strings.
	c := make(chan Message)
	go func() { // We launch the goroutine from inside the function.
		for i := 0; ; i++ {
			c <- Message{fmt.Sprintf("%s: %d", msg, i), waitForIt}
			time.Sleep(time.Duration(rand.Intn(2e3)) * time.Millisecond)
			//这里会堵塞，等待消费者先消费先
			// Each speaker must wait for a go-ahead.
			<-waitForIt
		}
	}()
	return c // Return the channel to the caller.
}

func fanIn4(input1, input2 <-chan Message) <-chan Message {
	c := make(chan Message)
	go func() {
		for {
			c <- <-input1
		}
	}()
	go func() {
		for {
			c <- <-input2
		}
	}()
	return c
}

// Generator: function that returns a channel
func boring6(msg string, quit chan bool) <-chan string { // Returns receive-only channel of strings.
	c := make(chan string)
	go func() { // We launch the goroutine from inside the function.
		for i := 0; ; i++ {
			select {
			case c <- fmt.Sprintf("%s %d", msg, i):

			}

			//time.Sleep(time.Duration(rand.Intn(1e3)) * time.Millisecond)

		}
	}()
	return c // Return the channel to the caller.
}

func main() {
	if false {

	}
	if false {
		// The functionality is analogous to the & on the end of a shell command.
		// 这种main都结束了，必须用&这种守护进程的方式去执行
		go boring1("boring!")
		fmt.Println("I'm listening.")
		time.Sleep(2 * time.Second)
		fmt.Println("You're boring; I'm leaving.")
	}
	if false {
		//外层建立ch，传递给发送方使用
		//Using channels
		c := make(chan string)
		go boring2("boring!", c)
		for i := 0; i < 5; i++ {
			fmt.Printf("You say: %q\n", <-c) // Receive expression is just a value.
		}
		fmt.Println("You're boring; I'm leaving.")
	}

	if false {
		// 发送方构建ch，返回ch，让接收方消费
		// Channels are first-class values, just like strings or integers.
		// Channels are first-class values, just like strings or integers.
		// Channels are first-class values, just like strings or integers.
		c := boring3("boring!") // Function returning a channel.
		for i := 0; i < 5; i++ {
			fmt.Printf("You say: %q\n", <-c)
		}
		fmt.Println("You're boring; I'm leaving.")
	}

	if false { //Channels as a handle on a service
		//Our boring function returns a channel that lets us communicate with the boring service it provides.
		joe := boring3("Joe")
		ann := boring3("Ann")
		for i := 0; i < 5; i++ {
			fmt.Println(<-joe)
			fmt.Println(<-ann)
		}
		fmt.Println("You're both boring; I'm leaving.")
	}

	if false {
		//Multiplexing
		//These programs make Joe and Ann count in lockstep.
		c := fanIn(boring3("Joe"), boring3("Ann"))
		for i := 0; i < 10; i++ {
			fmt.Println(<-c)
		}
		fmt.Println("You're both boring; I'm leaving.")
	}
	if false {
		/*
			Restoring sequencing
			Send a channel on a channel, making goroutine wait its turn.
			在channel上发送一个channel，只不过这个channel包裹在Message结构体里面

		*/
		waitForIt := make(chan bool) // Shared between all messages.
		c := fanIn4(boring4("Joe", waitForIt), boring4("Ann", waitForIt))
		for i := 0; i < 5; i++ {
			msg1 := <-c
			fmt.Println(msg1.str)
			msg2 := <-c
			fmt.Println(msg2.str)
			//Receive all messages, then enable them again by sending on a private channel
			msg1.wait <- true
			msg2.wait <- true
		}
	}

	if false {
		//Timeout using select
		c := boring3("Joe")
		for {
			select {
			case s := <-c:
				fmt.Println(s)
				// a timeout for each message.
			case <-time.After(1 * time.Second):
				fmt.Println("You're too slow.")
				return
			}
		}
	}

	if false {
		//Timeout for whole conversation using select
		c := boring3("Joe")
		//Create the timer once, outside the loop, to time out the entire conversation.
		timeout := time.After(3 * time.Second)
		for {
			select {
			case s := <-c:
				fmt.Println(s)
			case <-timeout:
				fmt.Println("You talk too much.")
				return
			}
		}
	}

	if true {
		// Quit channel
		quit := make(chan bool)
		c := boring("Joe", quit)
		for i := rand.Intn(10); i >= 0; i-- {
			fmt.Println(<-c)
		}
		quit <- true
	}
}
