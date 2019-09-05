package main

import (
	"os"
	"runtime/debug"
	"runtime/pprof"
	"time"
)

/*
调试利器：dump goroutine 的 stacktrace
https://colobu.com/2016/12/21/how-to-dump-goroutine-stack-traces/

[译]Go Stack Trace
https://colobu.com/2016/04/19/Stack-Traces-In-Go/

*/

func main() {
	go a()
	m1()
}

func m1() {
	m2()
}

func m2() {
	m3()
}

func m3() {
	if false {
		panic("panic from m3")
	}
	if false { //打印出当前goroutine的 stacktrace
		debug.PrintStack()
		time.Sleep(time.Hour)
		/*
			goroutine 1 [running]:
			runtime/debug.Stack(0xc04208202f, 0x0, 0xc042048180)
				d:/Go/src/runtime/debug/stack.go:24 +0xae
			runtime/debug.PrintStack()
				d:/Go/src/runtime/debug/stack.go:16 +0x29
			main.m3()
				D:/gitPro/go/src/demo/traceStack/traceStack.go:26 +0x29
			main.m2()
				D:/gitPro/go/src/demo/traceStack/traceStack.go:18 +0x27
			main.m1()
				D:/gitPro/go/src/demo/traceStack/traceStack.go:14 +0x27
			main.main()
				D:/gitPro/go/src/demo/traceStack/traceStack.go:10 +0x41

		*/
	}
	if false { //打印出所有goroutine的 stacktrace
		pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
		time.Sleep(time.Hour)
		/*

			goroutine profile: total 2
			1 @ 0x42d021 0x42d115 0x444ad5 0x4c7a57 0x4537d1
			#	0x444ad4	time.Sleep+0x144	d:/Go/src/runtime/time.go:65
			#	0x4c7a56	main.a+0x36		D:/gitPro/go/src/demo/traceStack/traceStack.go:54

			1 @ 0x4c0cd9 0x4c0ac7 0x4bd5e2 0x4c79ec 0x4c7967 0x4c7927 0x4c78e1 0x42cb92 0x4537d1
			#	0x4c0cd8	runtime/pprof.writeRuntimeProfile+0xa8	d:/Go/src/runtime/pprof/pprof.go:637
			#	0x4c0ac6	runtime/pprof.writeGoroutine+0xa6	d:/Go/src/runtime/pprof/pprof.go:599
			#	0x4bd5e1	runtime/pprof.(*Profile).WriteTo+0x3b1	d:/Go/src/runtime/pprof/pprof.go:310
			#	0x4c79eb	main.m3+0x6b				D:/gitPro/go/src/demo/traceStack/traceStack.go:48
			#	0x4c7966	main.m2+0x26				D:/gitPro/go/src/demo/traceStack/traceStack.go:20
			#	0x4c7926	main.m1+0x26				D:/gitPro/go/src/demo/traceStack/traceStack.go:16
			#	0x4c78e0	main.main+0x40				D:/gitPro/go/src/demo/traceStack/traceStack.go:12
			#	0x42cb91	runtime.main+0x221			d:/Go/src/runtime/proc.go:195


		*/
	}
}

func a() {
	time.Sleep(time.Hour)
}

/*
你可以使用runtime.Stack得到所有的goroutine的stack trace信息，事实上前面debug.PrintStack()也是通过这个方法获得的。
*/
func DumpStacks() {
	buf := make([]byte, 16384)
	buf = buf[:runtime.Stack(buf, true)]
	fmt.Printf("=== BEGIN goroutine stack dump ===\n%s\n=== END goroutine stack dump ===", buf)
}

/*
如果你的代码中配置了 http/pprof,你可以通过下面的地址访问所有的groutine的堆栈：
http://localhost:8888/debug/pprof/goroutine?debug=2.

*/

/*
panic: panic from m3

goroutine 1 [running]:
main.m3()
	D:/gitPro/go/src/demo/traceStack/traceStack.go:21 +0x40
main.m2()
	D:/gitPro/go/src/demo/traceStack/traceStack.go:17 +0x27
main.m1()
	D:/gitPro/go/src/demo/traceStack/traceStack.go:13 +0x27
main.main()
	D:/gitPro/go/src/demo/traceStack/traceStack.go:9 +0x41

从这个信息中我们可以看到p.go的第9行是main方法内，它在这一行调用m1方法，
m1方法在第13行调用m2方法，m2方法在第17行调用m3方法，m3方法在第21出现panic，
它们运行在goroutine 1中，当前goroutine 1的状态是running状态。


如果想让它把所有的goroutine信息都输出出来，可以设置 GOTRACEBACK=1:
GOTRACEBACK=1 go run p.go
panic: panic from m3
goroutine 1 [running]:
panic(0x596a0, 0xc42000a1b0)
	/usr/local/Cellar/go/1.7.4/libexec/src/runtime/panic.go:500 +0x1a1
main.m3()
	/Users/yuepan/go/src/github.com/smallnest/dump/p.go:21 +0x6d
main.m2()
	/Users/yuepan/go/src/github.com/smallnest/dump/p.go:17 +0x14
main.m1()
	/Users/yuepan/go/src/github.com/smallnest/dump/p.go:13 +0x14
main.main()
	/Users/yuepan/go/src/github.com/smallnest/dump/p.go:9 +0x3a

goroutine 4 [sleep]:
time.Sleep(0x34630b8a000)
	/usr/local/Cellar/go/1.7.4/libexec/src/runtime/time.go:59 +0xe1
main.a()
	/Users/yuepan/go/src/github.com/smallnest/dump/p.go:25 +0x30
created by main.main
	/Users/yuepan/go/src/github.com/smallnest/dump/p.go:8 +0x35
exit status 2

同样你也可以分析这个stack trace的信息，得到方法调用点的情况，同时这个信息将两个goroutine的stack trace都打印出来了，而且goroutine 4的状态是sleep状态。
*/
