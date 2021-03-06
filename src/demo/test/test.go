package main

import "os"
import "fmt"
import "path"
import "net/url"

//import "demo/tt"
import "time"
import "strconv"
import "container/list"
import "runtime"
import "math/rand"
import "context"
import "unsafe"
import "strings"
import "bytes"

//import "math/rand"
//import "sync/atomic"

type PostParam struct {
	Cid    int64 `json:"tid"`
	Sid    int64 `json:"sid"`
	Uid    int64 `json:"uid"`
	Extend map[string]string
}

type student struct {
	Name  string
	Age   int
	Param []PostParam
}

func ff() (*int, error) {
	var b int = 3
	//return &b, nil
	return &b, fmt.Errorf("TEST")
}

func hole1() (err error) {
	var a *int
	if a, err := ff(); err != nil {
		fmt.Printf("%+v", err)
	} else {
		fmt.Printf("inside a: %+v\n", *a)
	}
	fmt.Printf("outsize a: %+v\n", a)
	return
}

func hole2() error {
	var err error
	var a *int
	if a, err := ff(); err != nil { //a和err都是if的局部变量，屏蔽了外层的err,a
		//if a, err = ff(); err != nil {
		fmt.Printf("%+v\n", err)
	} else {
		fmt.Printf("inside a: %+v\n", *a)
	}
	fmt.Printf("outsize a: %+v\n", a)
	fmt.Printf("outsize err: %+v\n", err)
	if a != nil {
		fmt.Printf("outsize a=: %+v\n", *a)
	}
	return err
}

type persistConn struct {
	broken bool // an error has happened on this connection; marked broken so it's not reused.
	reused bool
}

func ap(a []int) {
	a = append(a, 10)
}

func ffff() (string, error) {
	return "test scope of variable", nil
}

// 可变长参数
func sum(nums ...int) {
	fmt.Print(nums, " ")
	total := 0
	for _, num := range nums {
		total += num
	}
	fmt.Println(total)
}

func wrapSum(nums ...int) {
	sum(nums...)
}

func wrapSum2(nums []int) {
	sum(nums...)
}

func f123(bbool *bool) {
	*bbool = true
}

func main() {
	if false {
		fmt.Print(strings.TrimLeft(" \t\n  \r 123Hello, 111111", "\r\t\n "))
		fmt.Print(strings.TrimLeft("¡¡¡!!!Hello, Gophers!!!", "!¡"))
		fmt.Println(strings.Replace("oinkkk oink oink", "k", "ky", 10))
		fmt.Println(strings.TrimSpace(" \t\n Hello, Gophers \n\t\r\n123"))
	}

	if false {
		bbool := false
		f123(&bbool)
		fmt.Println("b=", bbool)
	}
	if false { // slice len 清0 cap不变
		letters := []string{"a", "b", "c", "d"}
		fmt.Println(cap(letters))
		fmt.Println(len(letters))
		// clear the slice
		letters = letters[:0]
		fmt.Println(cap(letters))
		fmt.Println(len(letters))
	}

	if false {
		var b bytes.Buffer
		str1 := "this is a first string"

		str2 := " this is a second string"

		b.WriteString(str1)

		b.WriteString(str2)

		str3 := b.String()
		fmt.Printf("%+v\n", str3)
		b.Truncate(10)
		fmt.Printf("%+v\n", b.String())
		//i := 0
		j := 0
		//j = i++
		fmt.Println(j)
	}

	if false {
		failLogName := fmt.Sprintf("fail.%d.txt", time.Now().Unix())
		file, err := os.OpenFile(failLogName, os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		data := "/b/core.jpg"
		file.WriteString(data)
		file.WriteString("\n")
		file.WriteString(data)
	}

	if false {
		var b *bool
		if b == nil {
			fmt.Println("b=null")
		}
		b = new(bool)
		//*b = true
		if b == nil {
			fmt.Println("b=null")
		} else {
			fmt.Println("b=", *b)
		}

		//b = nil
		//
		//
	}

	if false { // 可变长参数
		sum(1, 2, 3)
		nums := []int{1, 2, 3, 4}
		sum(nums...)
		wrapSum(1, 2, 3, 4, 5)
		wrapSum2(nums)
	}

	runtime.GOMAXPROCS(8)
	if false {
		println(20010 | 134217728)
	}

	if false {
		var name string
		//name := "HI"
		if name, err := ffff(); nil == err {
			println(name) //符合预期，预期是 test scope of variable
		}
		println(name) //空白，不合符预期，预期是 test scope of variable
	}

	if false {
		a := []int{}
		a = append(a, 7, 8, 9)
		fmt.Printf("len: %d cap:%d data:%+v\n", len(a), cap(a), a)
		ap(a)
		fmt.Printf("len: %d cap:%d data:%+v\n", len(a), cap(a), a)
		p := unsafe.Pointer(&a[2])
		q := uintptr(p) + 8
		t := (*int)(unsafe.Pointer(q))
		fmt.Println(*t)
	}

	if false {
		a := []int{7, 8, 9}
		fmt.Printf("len: %d cap:%d data:%+v\n", len(a), cap(a), a)
		ap(a)
		fmt.Printf("len: %d cap:%d data:%+v\n", len(a), cap(a), a)
		p := unsafe.Pointer(&a[2])
		q := uintptr(p) + 8
		t := (*int)(unsafe.Pointer(q))
		fmt.Println(*t)
	}

	if false {
		idleConnCh := make(map[string]chan *persistConn)
		key := "123"
		waitingDialer := idleConnCh[key]
		fmt.Printf("%+v\n", waitingDialer) //访问不存在的元素 为值类型的空值nil,而且map不会新增元素
		fmt.Printf("%#v\n", idleConnCh)
	}

	if false {
		//		hole1()
		hole2()
	}

	if false {
		var p *PostParam = &PostParam{
			Cid: 2,
			Uid: 1,
		}
		p.Extend["123"] = "234"
		fmt.Printf("%+v", p)

	}

	//	for循环的坑
	if false {
		var stus []*student
		stus = []*student{
			&student{Name: "one", Age: 18},
			&student{Name: "two", Age: 19},
		}
		data := make(map[int]*student)
		for i, v := range stus {
			v.Age = 20
			data[i] = v //应该改为：data[i] = &stus[i]
		}
		for i, v := range data {

			fmt.Printf("key=%d, value=%v \n", i, *v)
		}
	}

	if false {
		//		var res = make(map[int64][]*PostParam)
		var stus []*student
		stus = []*student{
			&student{Name: "one", Age: 18},
			&student{Name: "two", Age: 19},
			nil,
		}
		size := len(stus) - 3
		var channel *student = &student{
			Name:  "one",
			Age:   18,
			Param: make([]PostParam, 0, size),
		}
		var a *PostParam = &PostParam{}
		channel.Param = append(channel.Param, *a)
		channel.Param = append(channel.Param, *a)

		fmt.Printf("%+v\n", *channel)

	}

	if false {
		type favContextKey string

		f := func(ctx context.Context, k favContextKey) {
			if v := ctx.Value(k); v != nil {
				fmt.Println("found value:", v)
				return
			}
			fmt.Println("key not found:", k)
		}

		k := favContextKey("language")
		ctx := context.WithValue(context.Background(), k, "Go")

		f(ctx, k)
		f(ctx, favContextKey("color"))

	}

	if false {
		var s bool = false
		fmt.Printf("%v", s)
		var err error
		err = nil
		fmt.Printf("err:%v", err)
	}
	if true {
		//fmt.Println(time.Now().UnixNano())
		fmt.Println(time.Now().Unix())
		fmt.Println(time.Now().Hour())
		fmt.Println(time.Now().Minute())

	}

	if false {
		uri := (1<<8 | 26)
		fmt.Println("uri1=", uri)
		tt := 1 << 8
		res := tt | 26
		fmt.Println("uri2=", res)
		re := 2 | 16
		fmt.Println(re)

	}

	if false {
		//这就是坑
		fmt.Printf("size=%v\n", len([]string{"", "", ""})) //3
	}

	if false {
		rand.Seed(time.Now().Unix())
		//在没有seed的情况下，每次执行idx=1
		idx := rand.Intn(4)
		fmt.Println("idx=", idx)
	}

	if false { // "container/list"
		//idle list.List
		// Create a new list and put some numbers in it.
		l := list.New()
		e4 := l.PushBack(4)
		e1 := l.PushFront(1)
		l.InsertBefore(3, e4)
		l.InsertAfter(2, e1)

		// Iterate through list and print its contents.
		for e := l.Front(); e != nil; e = e.Next() {
			fmt.Println(e.Value)
		}

	}

	if false {
		//ss := tt.GetFileName()
		//fmt.Println(ss)
	}

	if false {
		fmt.Println("args=", os.Args[0])
		process_name := path.Base(os.Args[0])
		fmt.Println("p=", process_name)
		/*
			Linux 输出
				args= ./test
				p= test
			Windows输出
				args= D:\SvnRepo\go\src\demo\test\test.exe
				p= D:\SvnRepo\go\src\demo\test\test.exe
		*/
	}

	if false { //构造Get请求参数的另一种方式
		u, _ := url.Parse("http://localhost:9001/xiaoyue")
		q := u.Query()
		q.Set("username", "user")
		q.Set("password", "passwd")
		u.RawQuery = q.Encode()
		fmt.Println(u.String())
	}

	if false {
		//Daisy-chain   相当于一个链式处理器
		//https://talks.golang.org/2012/concurrency.slide#39
		const n = 10000
		leftmost := make(chan int)
		right := leftmost
		left := leftmost
		for i := 0; i < n; i++ {
			right = make(chan int)
			go f(left, right)
			left = right
		}
		go func(c chan int) {
			c <- 1
		}(right)
		fmt.Println(<-leftmost)
	}
	if false { //时间差
		s2 := time.Now()
		s1 := time.Now().UnixNano()
		sum := int64(123)
		str := ""
		var i int64
		for i = 0; i < 800; i++ {
			sum += i
			str += fmt.Sprintf("%s", strconv.FormatInt(sum, 10))
			//str += "hi"
		}
		d2 := time.Since(s2)
		s3 := time.Now().UnixNano()
		d1 := s3 - s1
		//windows 下无法精确到微秒级别，最小的都是毫秒
		fmt.Println("len=", len(str), "ns=", d2.Nanoseconds(), "s1=", s1, "s3=", s3, "d1=", d1)
	}

	fmt.Println("hello123")
}

func f(left, right chan int) {
	//left <- 1 + <-right
	left <- (1 + <-right)
}
