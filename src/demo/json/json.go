package main

import "fmt"

//import "path"
//import "net/url"
import "encoding/json"
import "bytes"

//import "demo/test/tt"
//import "time"
//import "strconv"
//import "container/list"
//import "runtime"
//import "math/rand"
//import "context"

//import "math/rand"
//import "sync/atomic"

// https://stackoverflow.com/questions/38977555/error-embedding-n-into-a-string-literal
// json unmarshal 报错  invalid character '\n' in string literal

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

type AppManagerResp struct {
	Status    string      `json:"status"`
	SuccssIds []int64     `json:"succssids"`
	Expand    interface{} `json:"expand"`
}

type Data struct {
	Val string `json:"val"`
}

type Data1 struct {
	Val string `json:"val"`
}

func (t *Data1) JSON() ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}

//不会强行转义的编码函数，但是末尾会有换行符
func JSONMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}

//不会强行转义的编码函数，且末尾没有换行符
func JSONMarshal2(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	if buffer.Len() >= 1 {
		buffer.Truncate(buffer.Len() - 1)
	}
	return buffer.Bytes(), err
}

func main() {

	if false { // 字符串含有&时marshal乱码
		origin := &Data{
			Val: "Hello&&&World",
		}
		b, _ := json.Marshal(origin)
		fmt.Printf("%+v\n", string(b))

	}

	if true { // 字符串含有&时marshal乱码 解决方案  自定义编码器
		//https://stackoverflow.com/questions/28595664/how-to-stop-json-marshal-from-escaping-and
		origin := &Data1{
			Val: "Hello&&&World",
		}
		b, err := origin.JSON()
		fmt.Printf("b: %+v   err: %+v\n", string(b), err)

		dataJson := make(map[string]string)
		dataJson["pack_url"] = "Hello&&&World"
		a, err := JSONMarshal2(dataJson)
		fmt.Printf("a: %+v123\n len: %d", string(a), len(a)) //a: {"pack_url":"Hello\u0026\u0026\u0026World"}

	}

	if false { // RawMessage
		// RawMessage is a raw encoded JSON value.
		// It implements Marshaler and Unmarshaler and can be used to delay JSON decoding or precompute a JSON encoding.
		// https://golang.org/pkg/encoding/json/#RawMessage
		// h是预先组装好的json， 定义为RawMessage，方便marshal时不会追加额外的/
		h := json.RawMessage(`{"precomputed": true}`)
		// h1作为参照，定义为普通的string，结果最后被加了/
		h1 := `{"precomputed": true}`
		//h1 := "{\"precomputed\": true}"

		c := struct {
			Header  *json.RawMessage `json:"header"`
			Header1 string           `json:"header1"`
			Body    string           `json:"body"`
		}{Header: &h, Header1: h1, Body: "Hello Gophers!"}

		b, err := json.MarshalIndent(&c, "", "  ")
		//b, err := json.Marshal(&c)
		if err != nil {
			fmt.Println("error:", err)
		}
		fmt.Printf("b: %+v   err: %+v\n", string(b), err)
		/*  可以看到 header1自己添加了反斜杠\
				b: {
		  "header": {
		    "precomputed": true
		  },
		  "header1": "{\"precomputed\": true}",
		  "body": "Hello Gophers!"
		}   err: <nil>
		*/
	}

	if false { // 模拟json解码\n报错

		//data := `{"val":"HelloWorld"}`
		// OK

		//data := `{"val":"Hello\nWorld"}`
		// OK

		//data := "{\"val\":\"Hello\nWorld\"}"
		// 报错 UnMarshal Err: invalid character '\n' in string literal
		// 也就是如果有\n 必须用``抱起来

		origin := &Data{
			Val: "Hello\r\nWorld",
		}
		// 如果元数据就有\n marshal不会报错，unmarshal也不会报错
		b, _ := json.Marshal(origin)

		resp := &Data{}
		//err := json.Unmarshal([]byte(data), resp)
		err := json.Unmarshal(b, resp)
		if err != nil {
			fmt.Printf("UnMarshal Err: %+v\n", err)
		} else {
			fmt.Printf("%+v\n", resp)
		}
	}

	if false {
		// 无status字段，则反解时，status为空字符串
		//data := `{"appid":30032,"expand":"{\"req_chn\":\"216172782113791086\"}","seq":2147483646,"sid":1350228370,"subsidapp":1,"succssids":[1350228370]}`
		// 当expand为object时，不能直接反序列化为string
		data := `{"appid":30032,"expan":{"req_chn":"216172782113791086", "name":"高清", "gear":1},"seq":2147483646,"sid":1350228370,"succssids":[1350228370]}`
		var resp AppManagerResp
		err := json.Unmarshal([]byte(data), &resp)
		if err != nil {
			fmt.Printf("UnMarshal Err: %+v\n", err)
		} else {
			fmt.Printf("%+v\n", resp)
			fmt.Println("status=", resp.Status)
			fmt.Println(resp.Status == "")
		}
		//Expand这个interface在反序列化时变成了实际类型为map  Expand:map[req_chn:216172782113791086 name:高清 gear:1]}
		// 直接把这个map打包成 json字符串
		fmt.Printf("%+v\n", resp.Expand)
		exp, _ := json.Marshal(resp.Expand)
		fmt.Printf("%T exp: %+v \n", string(exp), string(exp))
	}

	if false {
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
			fmt.Printf("DecodeMsg: %#v\n", deMsg)
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
	}
	fmt.Println("------hello world--------")
}
