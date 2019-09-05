package main

import (
	"fmt"
	"time"
)

func main() {

	// map safe
	if false {
		//map不是协程(goroutine)安全的
		m := make(map[int]int)
		go func() {
			for {
				_ = m[1]
			}
		}()
		go func() {
			for {
				m[2] = 2
			}
		}()
		select {}
		//报错  fatal error: concurrent map read and map write
	}

	// var safe
	if true {
		//并发不安全，多个协程对同一个变量进行读写操作。所以需要原子操作来保证线程安全.
		var cnt uint32 = 0
		for i := 0; i < 10; i++ {
			go func() {
				for i := 0; i < 20; i++ {
					time.Sleep(time.Millisecond)
					//atomic.AddUint32(&cnt, 1)
					cnt = cnt + 1
				}
			}()
		}
		time.Sleep(time.Second) //等一秒钟等goroutine完成
		//cntFinal := atomic.LoadUint32(&cnt) //取数据
		//fmt.Println("cnt:", cntFinal)

		fmt.Println("cnt:", cnt)
		//cnt: 199
	}
}
