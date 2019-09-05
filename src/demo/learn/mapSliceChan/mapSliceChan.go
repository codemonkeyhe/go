package main

import (
	"fmt"
	"time"
)

type Param struct {
	Cid    int64 `json:"tid"`
	Sid    int64 `json:"sid"`
	Uid    int64 `json:"uid"`
	Extend map[string]string
}

func add(a *[]int) {
	*a = append(*a, 3)
	fmt.Println("len=", len(*a), "cap=", cap(*a))
	*a = append(*a, 4)
	fmt.Println("len=", len(*a), "cap=", cap(*a))
	*a = append(*a, 5)
	fmt.Println("len=", len(*a), "cap=", cap(*a))
}

//此arr是副本，此arr.len的变化不会影响到外部的实参arr.len，但是此arr.ptr和实参arr.ptr的指向是相同的
func getParam(arr []int, md map[int]*Param, md2 map[int]Param) {
	fmt.Printf("len=%d cap=%d\n", len(arr), cap(arr))
	arr = append(arr, 1)
	fmt.Printf("len=%d cap=%d\n", len(arr), cap(arr))
	md[3] = &Param{
		Cid: 123,
		Uid: 456,
	}
	md2[11] = Param{
		Cid: 111,
		Uid: 444,
	}
}

var globalArr = [2]string{"HI", "Glo"}

var gS = &Param{
	Cid: 123,
	Sid: 123,
}
var gm = map[int]int{1: 2, 3: 4}

func main() {
	if true { //全局结构体
		fmt.Println(globalArr)
		fmt.Println(gS)
		fmt.Println(gm)
	}

	if false {
		a := make([]int, 32)
		fmt.Printf("len=%d cap=%d\n", len(a), cap(a))
		a = append(a, 1)
		fmt.Printf("len=%d cap=%d\n", len(a), cap(a))
	}

	//slice作为函数参数
	//slice底层只是ptr指向的数据共享，每个slice自己的len和cap是不共享的
	if false {
		arr := make([]int, 0, 4)
		fmt.Printf("len=%d cap=%d\n", len(arr), cap(arr))
		md := make(map[int]*Param)
		md2 := make(map[int]Param)
		getParam(arr, md, md2)
		fmt.Printf("len=%d cap=%d\n", len(arr), cap(arr))
		fmt.Printf("arr=%+v\n", arr)
		fmt.Printf("md=%+v\n", md)
		fmt.Printf("md2=%+v\n", md2)
	}

	if false {
		if true {
			a := make([]int, 0)
			fmt.Printf("len=%d cap=%d\n", len(a), cap(a)) // len=0 cap=0
			b := append(a, 1)
			fmt.Printf("len=%d cap=%d\n", len(a), cap(a)) // len=0 cap=0  对a执行append并没有改为a的len，
			fmt.Printf("len=%d cap=%d\n", len(b), cap(b))
			_ = append(a, 2) // 因为a的cap=0，所以相当于新分配了内存
			fmt.Printf("len=%d cap=%d\n", len(a), cap(a))
			println(b[0]) // 1
		} else {
			a := make([]int, 0, 10)
			fmt.Printf("len=%d cap=%d\n", len(a), cap(a))
			b := append(a, 1)
			fmt.Printf("len=%d cap=%d\n", len(a), cap(a))
			fmt.Printf("len=%d cap=%d\n", len(b), cap(b))
			_ = append(a, 2) // 因为a的cap=10，所以没有新分配了内存， 又因为len(a)=0，相当于a[0]=2
			fmt.Printf("len=%d cap=%d\n", len(a), cap(a))
			println(b[0]) // 2
		}

	}

	// 多余写法
	if false {
		res := make(map[int64][]*Param)
		p := &Param{
			Cid: 0,
			Sid: 1,
			Uid: 2,
		}
		res[123] = append(res[123], p)
		res[123] = append(res[123], p)
		fmt.Println(res)

		//下面的写法比较多余
		if false {
			if list, ok := res[234]; !ok {
				res[234] = make([]*Param, 0)
				//错误写法，不存在是不可以使用list   list = append(list, p)
				res[234] = append(res[234], p)
			} else {
				list = append(list, p)
			}

		} else {
			//简洁写法
			res[234] = append(res[234], p)
		}
		fmt.Println(res)
	}

	//传递slice的指针
	if false {
		s := make([]int, 0, 0)
		fmt.Println("len=", len(s), "cap=", cap(s))
		add(&s)
		fmt.Printf("%+v\n", s)
		fmt.Println("len=", len(s), "cap=", cap(s))
	}

	// access unexist element in map
	if false {
		m1 := make(map[int32]int32)
		m1[123] = 1
		m1[456] = 1
		m1[789] += 2
		m1[123] += 2
		m1[555]++
		m1[555]++
		m1[555]++
		ss := m1[111]
		fmt.Printf("%+v ss:%d\n", m1, ss)
	}

	// range chan
	if false {
		ic := make(chan int)
		go func() {
			ic <- 1
			ic <- 2
			close(ic)
		}()
		go func() {
			for i := range ic {
				fmt.Println(i)
			}
		}()
		//wait go routine to run
		time.Sleep(time.Second * 1)
	}

}
