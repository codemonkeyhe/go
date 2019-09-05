package main

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
)

/*
https://zhuanlan.zhihu.com/p/51630005

https://segmentfault.com/a/1190000009133154
*/

func main() {
	// 用 chan func() 传递需要执行的函数
	ch := make(chan func())
	// 开启多个 goroutine
	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			// 循环读取 chan func() 并执行
			for fn := range ch {
				fn()
			}
		}()
	}

	// 执行一个函数，不等待完成
	do := func(fn func()) {
		ch <- fn
	}

	// 执行一个函数，并等待完成
	syncDo := func(fn func()) {
		var l sync.Mutex
		l.Lock()
		ch <- func() {
			fn()
			l.Unlock()
		}
		l.Lock()
	}

	// 停止所有 goroutine
	stop := func() {
		close(ch)
	}

	// 一个用例
	var sum int64
	for i := 0; i < 10000000; i++ {
		i := i
		syncDo(func() {
			atomic.AddInt64(&sum, int64(i))
		})
		if i > 0 && i%1000000 == 0 {
			do(func() {
				fmt.Printf("%d\n", atomic.LoadInt64(&sum))
			})
		}
	}
	fmt.Printf("sum %d\n", sum)

	stop()

}
