package main

/*
https://colobu.com/2018/11/03/get-function-name-in-go/?hmsr=toutiao.io&utm_medium=toutiao.io&utm_source=toutiao.io
https://colobu.com/2016/12/21/how-to-dump-goroutine-stack-traces/

FuncForPC 是一个有趣的函数， 它可以把程序计数器地址对应的函数的信息获取出来。如果因为内联程序计数器对应多个函数，它返回最外面的函数。
它的返回值是一个*Func类型的值，通过*Func可以获得函数地址、文件行、函数名等信息。
除了上面获取程序计数器的方式，也可以通过反射的方式获取函数的地址：


runtime.FuncForPC(reflect.ValueOf(foo).Pointer()).Name()


*/
import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/petermattis/goid"
)

func main() {
	TestGetGoroutineId()
	Foo()
	print("Done\n")
}

func Foo() {
	fmt.Printf("我是 %s, %s 在调用我!\n", printMyName(), printCallerName())
	Bar()
}

func Bar() {
	fmt.Printf("我是 %s, %s 又在调用我!\n", printMyName(), printCallerName())
	trace()
	fmt.Printf("-----------\n")
	trace2()
	fmt.Printf("-----------\n")
	DumpStacks()
}

// 0 代表当前函数，也是调用runtime.Caller的函数。1 代表上一层调用者，以此类推。
func printMyName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}

func printCallerName() string {
	pc, _, _, _ := runtime.Caller(2)
	return runtime.FuncForPC(pc).Name()
}

/*
你可以通过runtime.Caller、runtime.Callers、runtime.FuncForPC等函数更详细的跟踪函数的调用堆栈。
func Caller(skip int) (pc uintptr, file string, line int, ok bool)
Caller可以返回函数调用栈的某一层的程序计数器、文件信息、行号。
0 代表当前函数，也是调用runtime.Caller的函数。1 代表上一层调用者，以此类推。


func Callers(skip int, pc []uintptr) int
Callers用来返回调用站的程序计数器, 放到一个uintptr中。
0 代表 Callers 本身，这和上面的Caller的参数的意义不一样，历史原因造成的。 1 才对应这上面的 0。
比如在上面的例子中增加一个trace函数，被函数Bar调用。
*/

func trace() {
	pc := make([]uintptr, 10) // at least 1 entry needed
	n := runtime.Callers(0, pc)
	for i := 0; i < n; i++ { //从最底层函数到main入口
		f := runtime.FuncForPC(pc[i])
		file, line := f.FileLine(pc[i])
		fmt.Printf("%s:%d %s\n", file, line, f.Name())
	}
}

/*
d:/Go/src/runtime/extern.go:212 runtime.Callers
D:/gitPro/go/src/demo/trace/trace.go:72 main.trace
D:/gitPro/go/src/demo/trace/trace.go:39 main.Bar
D:/gitPro/go/src/demo/trace/trace.go:35 main.Foo
D:/gitPro/go/src/demo/trace/trace.go:29 main.main
d:/Go/src/runtime/proc.go:204 runtime.main
d:/Go/src/runtime/asm_amd64.s:2338 runtime.goexit
*/

/*
上面的Callers只是或者栈的程序计数器，如果想获得整个栈的信息，可以使用CallersFrames函数，省去遍历调用FuncForPC。
上面的trace函数可以更改为下面的方式
*/
func trace2() {
	pc := make([]uintptr, 10) // at least 1 entry needed
	n := runtime.Callers(0, pc)
	frames := runtime.CallersFrames(pc[:n])
	for {
		frame, more := frames.Next()
		fmt.Printf("%s:%d %s\n", frame.File, frame.Line, frame.Function)
		if !more {
			break
		}
	}
}

/*
在程序panic的时候，一般会自动把堆栈打出来，如果你想在程序中获取堆栈信息，可以通过debug.PrintStack()打印出来。
比如你在程序中遇到一个Error,但是不期望程序panic,只是想把堆栈信息打印出来以便跟踪调试，你可以使用debug.PrintStack()。
*/
func DumpStacks() {
	buf := make([]byte, 16384)
	buf = buf[:runtime.Stack(buf, true)]
	fmt.Printf("=== BEGIN goroutine stack dump ===\n%s\n=== END goroutine stack dump ===", buf)
}

func TestGetGoroutineIdSlow() {
	fmt.Println("TestGetGoroutineIdSlow", GoID())
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Println("i=", i, "Gid=", GoID())
		}()
	}
	wg.Wait()
}

/*
利用堆栈信息还可以获取goroutine的id
https://colobu.com/2016/04/01/how-to-get-goroutine-id/
它利用runtime.Stack的堆栈信息。runtime.Stack(buf []byte, all bool) int会将当前的堆栈信息写入到一个slice中，
堆栈的第一行为goroutine #### […,其中####就是当前的gororutine id,通过这个花招就实现GoID方法了。
但是需要注意的是，获取堆栈信息会影响性能，所以建议你在debug的时候才用它
*/
func GoID() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

func TestGetGoroutineId() {
	fmt.Println("TestGetGoroutineId", goid.Get())
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Println("i=", i, "Gid=", goid.Get())
		}()
	}
	wg.Wait()
}
