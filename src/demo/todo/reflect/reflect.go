package main

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

/*
GO语言提供了一种机制，在编译时不知道类型的情况下，可更新变量、在运行时查看值、调用方法以及直接对它们的布局进行操作，这种机制称为反射。
*/

// Any formats any value as a string.
func formatAny(value interface{}) string {
	return formatAtom(reflect.ValueOf(value))
}

// formatAtom formats a value without inspecting its internal structure.
func formatAtom(v reflect.Value) string {
	switch v.Kind() {
	case reflect.Invalid: //表示没有任何值，reflect.Value的零值属于Invalid类型
		return "invalid"
	case reflect.Int, reflect.Int8, reflect.Int16, //基础类型
		reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, //基础类型
		reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(v.Uint(), 10)
	// ...floating-point and complex cases omitted for brevity...
	case reflect.Bool: //基础类型
		return strconv.FormatBool(v.Bool())
	case reflect.String: //基础类型
		return strconv.Quote(v.String())
	case reflect.Chan, reflect.Func, reflect.Ptr, reflect.Slice, reflect.Map: //引用类型
		return v.Type().String() + " 0x" +
			strconv.FormatUint(uint64(v.Pointer()), 16)
	default: // reflect.Array, reflect.Struct, reflect.Interface	//聚合类型 和 接口类型
		//这个分支处理的不够完善
		return v.Type().String() + " value"
	}
}

/*
type Kind uint

const (
	Invalid Kind = iota
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Uintptr
	Float32
	Float64
	Complex64
	Complex128
	Array
	Chan
	Func
	Interface
	Map
	Ptr
	Slice
	String
	Struct
	UnsafePointer
)
*/

func TestFormatAny() {
	// The pointer values are just examples, and may vary from run to run.
	//!+time
	var x int64 = 1
	var d time.Duration = 1 * time.Nanosecond
	fmt.Println(formatAny(x))                  // "1"
	fmt.Println(formatAny(d))                  // "1"
	fmt.Println(formatAny([]int64{x}))         // "[]int64 0x8202b87b0"
	fmt.Println(formatAny([]time.Duration{d})) // "[]time.Duration 0x8202b87e0"
	fmt.Println(formatAny(nil))                // invalid
	fmt.Println(formatAny(false))              // false
	fmt.Println(formatAny([2]int{1, 2}))       // [2]int value
	//!-time
}

func Display(name string, x interface{}) {
	fmt.Printf("Display %s (%T):\n", name, x)
	display(name, reflect.ValueOf(x))
}

///////////////////////////////////////////////////////////////////////////!-Display

//!+display
func display(path string, v reflect.Value) {
	switch v.Kind() {
	case reflect.Invalid:
		fmt.Printf("%s = invalid\n", path)
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			display(fmt.Sprintf("%s[%d]", path, i), v.Index(i))
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			fieldPath := fmt.Sprintf("%s.%s", path, v.Type().Field(i).Name)
			display(fieldPath, v.Field(i))
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			display(fmt.Sprintf("%s[%s]", path,
				formatAtom(key)), v.MapIndex(key))
		}
	case reflect.Ptr:
		if v.IsNil() {
			fmt.Printf("%s = nil\n", path)
		} else {
			display(fmt.Sprintf("(*%s)", path), v.Elem())
		}
	case reflect.Interface:
		if v.IsNil() {
			fmt.Printf("%s = nil\n", path)
		} else {
			fmt.Printf("%s.type = %s\n", path, v.Elem().Type())
			display(path+".value", v.Elem())
		}
	default: // basic types, channels, funcs
		fmt.Printf("%s = %s\n", path, formatAtom(v))
	}
}

func main() {
	fmt.Println("Hi")

	TestFormatAny()

	if false {
		// func TypeOf(i interface{}) Type
		t := reflect.TypeOf(3)
		fmt.Println(t)          //int
		fmt.Printf("t=%v\n", t) //int
		fmt.Println(t.String()) // int
		//	func ValueOf(i interface{}) Value
		v := reflect.ValueOf(3)
		fmt.Println(v)          //3
		fmt.Printf("v=%v\n", v) //3  fmt的%v对reflect.Value进行了特殊的处理
		fmt.Println(v.String()) //<int Value>

		vs := reflect.ValueOf("123")
		fmt.Println(vs)           //1233
		fmt.Printf("vs=%v\n", vs) //vs=123  fmt的%v对reflect.Value进行了特殊的处理
		fmt.Println(vs.String())  //123		如果value包含的是字符串，则打印字符串本身，而不是打印类型
	}

	if false {
		v := reflect.ValueOf(3)
		//func (v Value) Interface() (i interface{})
		x := v.Interface()
		i := x.(int)
		fmt.Printf("%d\n", i)
	}

}
