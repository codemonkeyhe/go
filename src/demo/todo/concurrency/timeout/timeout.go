package main

import (
	"fmt"
	"math/rand"
	"time"
)

/*
Go Concurrency Patterns: Timing out, moving on
Andrew Gerrand
23 September 2010
https://blog.golang.org/go-concurrency-patterns-timing-out-and
*/

func main() {

	if false {
		TimeOut()
	}

	if true {
		cons := make([]Conn, 0, 4)
		cons = append(cons, Conn{id: 1})
		cons = append(cons, Conn{id: 2})
		cons = append(cons, Conn{id: 3})

		res := Query(cons, "cat")
		fmt.Println(res)
	}
}

func TimeOut() {
	ch := make(chan int)
	// 有缓冲的优点： 让发送方不阻塞，减少协程数目
	timeout := make(chan bool, 1)
	go func() {
		//https://stackoverflow.com/questions/37891280/goroutine-time-sleep-or-time-after
		//为了可读性，还是使用time.Sleep算了，两者差不多
		<-time.After(1 * time.Second)
		//time.Sleep(1 * time.Second)

		// The timeout channel is buffered with space for 1 value,
		// allowing the timeout goroutine to send to the channel and then exit
		// 发送方没有关闭timeout，GC会管理
		// The timeout channel will eventually be deallocated by the garbage collector.
		timeout <- true
	}()
	select {
	case <-ch:
		// a read from ch has occurred
	case <-timeout:
		fmt.Println("the read from ch has timed out")
		// the read from ch has timed out
	}
}

type Conn struct {
	id int
}

type Result string

func (p *Conn) DoQuery(key string) Result {
	rand.Seed(42)
	r := rand.Intn(5)
	time.Sleep(time.Second * time.Duration(r))
	return Result(fmt.Sprintf("id=%d Q=%s", p.id, key))
}

/*
Let's look at another variation of this pattern.
In this example we have a program that reads from multiple replicated databases simultaneously.
The program needs only one of the answers, and it should accept the answer that arrives first.

The function Query takes a slice of database connections and a query string.
It queries each of the databases in parallel and returns the first response it receives:








Question about Go concurrency example
https://groups.google.com/forum/#!topic/golang-nuts/4WJtV0hrXGY

All the DoQuery calls might take a long time to complete.
The channel send to ch only happens when a query has completed.
The first query that finishes will succeed in sending its result to the channel
(we know that for certain because the channel has a buffer size of 1).
That's the result that will be returned from Query.

The second query that finishes will also succeed in sending its result to the channel,
if it does so after the first result has been sent.
 however that result is never used.
All the other result sends will then block and fall through to the default clause accordingly.

Does that make things clearer?
*/
func Query(conns []Conn, query string) Result {
	//ch := make(chan Result)
	ch := make(chan Result, 1)
	for _, conn := range conns {
		go func(c Conn) {
			select {
			//先执行右侧的DoQuery，如果ch缓冲大小为1，那么第一个返回的结果将发送成功，然后被顺利消费后，Query函数就退出了。
			//其他的闭包协程往ch里面继续写结果时，第二个协程可以写成功，但是结果被丢弃了
			//其他协程写堵塞，因为select在case通信堵塞的情况下，执行default语句。所以，其他协程可以正常退出
			case ch <- c.DoQuery(query):
			default:
			}
		}(conn)
	}
	return <-ch
}
