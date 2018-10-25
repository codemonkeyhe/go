package main

import "fmt"

//import "path"
//import "net/url"
import "encoding/json"

//import "demo/test/tt"
//import "time"
//import "strconv"
//import "container/list"
//import "runtime"
//import "math/rand"
//import "context"

//import "math/rand"
//import "sync/atomic"

//只有字段名是大写的，才会被编码到json当中
// -忽略字符,1序列化到jsonstr忽略掉该字段，2. 从jsonstr反序列化时即使jsonstr中有同名的对象，也不会反序列化到结构体。
// 但是不能这样写 Ignore string `json:"ignore, -"` 这样的话 - 忽略符号则不生效

// omitempty 的作用是当一个field的值是empty的时候，序列化JSON时候忽略这个field 这里需要注意的是关于emtpty的定义：
// The empty values are false, 0, any nil pointer or interface value, and any array, slice, map, or string of length zero.
// omitempty只作用于序列化的过程，反序列化不起作用，而-即控制序列化，也控制反序列化。
type Message struct {
	Name    string `json:"name"`
	Body    string `json:"body"`
	Time    int64  `json:"ts,string,omitempty"`
	Ignore1 string `json:"-"`
	Ignore2 int64  `json:"-"`
	small   string `json:"small"`
	Source  int64  `json:"source"`
	Source1 *int64 `json:"s1"`
}

func main() {
	//结构体Marshal为jsonstr
	//s1 := int64(123)

	//m := Message{"Alice", "Hello", 1294706395881, "test", 404, "ddd", 1024, &s1}

	m := Message{"Alice", "Hello", 1294706395881, "test", 404, "ddd", 1024, nil}
	// {"name":"Alice","body":"Hello","ts":"1294706395881","source":1024,"s1":null}
	// 当 Source1 *int64 `json:"s1,omitempty"`  则为{"name":"Alice","body":"Hello","ts":"1294706395881","source":1024}

	//m := Message{"Alice", "Hello", 0, "test", 404, "ddd", 1024, nil}
	//{"name":"Alice","body":"Hello","source":1024,"s1":null}

	var b []byte
	var err error
	if b, err = json.Marshal(m); err != nil {
		fmt.Println("err: %v", err)
	} else {
		fmt.Println(string(b))
		//{"name":"Alice","body":"Hello"}  ts=0为被忽略了。 0认为是空值，所以被omitempty
	}

	//支持ts从string解码为Int64   结果为{Name:Alice Body:Hello Time:12947 Ignore1: Ignore2:0}
	//b = []byte(`{"name":"Alice","body":"Hello","ts":"12947"}`)

	//反序列化时，忽略掉Ingore1和Ingore2 结果为{Name:Alice Body:Hello Time:12947 Ignore1: Ignore2:0}
	//b = []byte(`{"name":"Alice","body":"Hello","ts":"12947", "Ignore1":"test", "Ignore2":123}`)

	//坑 虽然结果有Ignore2:0，但是不是来自jsonstr，而是默认值
	//b = []byte(`{"name":"Alice","body":"Hello","ts":"12947", "Ignore1":"test", "Ignore2":0}`)

	// body不存在时， deMsg.Body=="
	//b = []byte(`{"name":"Alice","ts":"12947"}`)
	//body存在时， deMsg.Body=="
	//b = []byte(`{"name":"Alice","body":"", "ts":"12947"}`)
	//也就是通过deMsg.Body=""无法判断jsonStr是否存在body这个key

	// source字段不存在时 反序列化后deMsg.Source==0
	//b = []byte(`{"name":"Alice","body":"Hello","ts":"1294","Ignore1":"111", "Ignore2":222,  "small":"ss"}`)
	// source字段存在且为0时 反序列化后deMsg.Source==0,所以无法通过deMsg.Source==0来判断源jsonstr是否存在source字段
	//b = []byte(`{"name":"Alice","body":"Hello","ts":"1294","Ignore1":"111", "Ignore2":222,  "small":"ss", "source":0}`)

	//b = []byte(`{"name":"Alice","body":"Hello","ts":"1294","Ignore1":"111", "Ignore2":222,  "small":"ss", "source":0, "s1":0}`)
	b = []byte(`{"name":"Alice","body":"Hello","ts":"1294","Ignore1":"111", "Ignore2":222,  "small":"ss", "source":0}`)

	var deMsg Message
	//jsonstr Unmarshal为结构体deMsg
	err = json.Unmarshal(b, &deMsg)
	if err != nil {
		fmt.Printf("JSON unmarshal Error: %s  jsonStr=%s\n", err, string(b))
	} else {
		fmt.Printf("%#v\n", deMsg)
		//字符串不会为nil
		if deMsg.Body == "" {
			fmt.Printf("NO BODY\n")
		}
		if deMsg.Source1 == nil {
			fmt.Printf("NO S1\n")
		} else {
			if *deMsg.Source1 == 0 {
				fmt.Printf("Exist S1 ,=0\n")
			}
		}

	}

	fmt.Println("hello world")
}
