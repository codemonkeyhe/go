package main

import "os"
import "fmt"
import "path"
import "net/url"
import "encoding/json"
import "demo/test/tt"
import "time"
import "strconv"
import "container/list"
import "runtime"
import "math/rand"
import "context"

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

func main() {
	runtime.GOMAXPROCS(8)
	if true {
		var p *PostParam = &PostParam{
			Cid: 2,
			Uid: 1,
		}
		p.Extend["123"] = "234"
		fmt.Printf("%+v", p)

	}

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
	if false {
		fmt.Println(time.Now().UnixNano())
	}

	if false {
		res := make(map[int64][]*PostParam)
		p := &PostParam{
			Cid: 0,
			Sid: 1,
			Uid: 2,
		}
		res[123] = append(res[123], p)
		res[123] = append(res[123], p)
		fmt.Println(res)
		//下面的写法比较多余
		if list, ok := res[234]; !ok {
			res[234] = make([]*PostParam, 0)
			//错误写法，不存在是不可以使用list   list = append(list, p)
			res[234] = append(res[234], p)
		} else {
			list = append(list, p)
		}
		fmt.Println(res)
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

	if false {
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
		ss := tt.GetFileName()
		fmt.Println(ss)
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

	if false { //解析json
		result := make([]*PostParam, 0)
		jsonstr := `[{"tid":61032445,"sid":2233637304,"uid":123456},{"tid":234,"sid":2345,"uid":654321}]`
		err := json.Unmarshal([]byte(jsonstr), &result)
		if err != nil {
			fmt.Printf("Err:%v\n", err)
		}
		fmt.Printf("Data:+%v\n", result[0])
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
