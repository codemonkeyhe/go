package main

/*
#cgo CFLAGS: -I./include
#cgo LDFLAGS: -L./lib -lfoo -Wl,-rpath,./lib/

#include "include/foo.h"

*/
import "C"

import (
	"fmt"
	"unsafe"
)

func main() {
	// 调用动态库函数fun1
	C.fun1()
	// 调用动态库函数fun2
	C.fun2(C.int(4))
	// 调用动态库函数fun3
	var pointer unsafe.Pointer
	ret := C.fun3(&pointer)
	fmt.Println(int(ret))
}
