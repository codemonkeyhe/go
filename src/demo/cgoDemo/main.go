package main

import (
	"fmt"
	"unsafe"
)

// #include <stdio.h>
// #include <stdlib.h>
/*
void print(char *str) {
	printf("%s\n", str);
}
int PlusOne(int n)
{
	return n + 1;
}
typedef struct _POINT
{
	double x;
	double y;
}POINT;
*/
import "C"

func main() {
	{ //变量demo
		var n C.int
		n = 5
		fmt.Println(n) // 5

		var m1 int
		// Go不认为C.int与int、int32等类型相同
		// 所以必须进行转换
		m1 = int(n + 3)
		fmt.Println(m1) // 8

		var m2 int32
		m2 = int32(n + 20)
		fmt.Println(m2) // 25
	}

	{ //函数demo

		var n int = 10
		var m int = int(C.PlusOne(C.int(n))) // 类型要转换
		fmt.Println(m)                       // 11
	}

	{ //结构体
		var p C.POINT
		p.x = 9.45
		p.y = 23.12
		fmt.Println(p) // {9.45 23.12}
	}

	{ //http://tonybai.com/2012/09/26/interoperability-between-go-and-c/
		s := "Hello Cgo"
		cs := C.CString(s)
		C.print(cs)
		C.free(unsafe.Pointer(cs))
	}
}
