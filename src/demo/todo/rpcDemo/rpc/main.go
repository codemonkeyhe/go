package main

import (
	"demo/rpc/server"
	//"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
)

/*
http://colobu.com/2016/09/18/go-net-rpc-guide/
如果bServer变量不是默认值，比如设置为ture，则会显示默认值为true
D:\SvnRepo\go\src\demo\rpc>rpc --help
Usage of rpc:
  -server
        Run server (default true)
如果bServer变量是默认值，比如设置为false，则不会显示默认值
D:\SvnRepo\go\src\demo\rpc>rpc --help
或者rpc -h
Usage of rpc:
  -server
        Run server

D:\SvnRepo\go\src\demo\rpc>rpc -server
run server...
REQ: Args:&{A:7 B:8}
RESP: reply:56
REQ: Args:&{A:7 B:8}
RESP: Quotient:{Quo:0 Rem:7}

D:\SvnRepo\go\src\demo\rpc>rpc
run client...
Arith: 7*8=56
RES: &{0 7}
replyCall: &{Arith.Divide 0xc042124100 0xc042124380 <nil> 0xc0421405a0}
divCall: &{Arith.Divide 0xc042124100 0xc042124380 <nil> 0xc0421405a0}


*/

func Usage() {
	fmt.Fprint(os.Stderr, "Usage of ", os.Args[0], ":\n")
	flag.PrintDefaults()
	fmt.Fprint(os.Stderr, "\n")
}

func main() {
	flag.Usage = Usage
	bServer := flag.Bool("server", false, "Run server")
	flag.Parse()
	if *bServer {
		fmt.Println("run server...")
		arith := new(server.Arith)

		rpc.Register(arith)

		rpc.HandleHTTP()
		l, e := net.Listen("tcp", ":1234")
		if e != nil {
			fmt.Printf("ERR: %v", e)
			log.Fatal("listen error:", e)
		}
		go http.Serve(l, nil)
		//阻塞，让服务端继续执行
		select {}
	} else {
		fmt.Println("run client...")
		serverAddress := "127.0.0.1"
		//客户端可以调用Dial和DialHTTP建立连接。
		client, err := rpc.DialHTTP("tcp", serverAddress+":1234")
		if err != nil {
			fmt.Printf("ERR: %v", err)
			log.Fatal("dialing:", err)
		}
		//同步调用
		args := &server.Args{7, 8}
		var reply int
		//Call是同步的方式调用，它实际是调用Go实现的，
		err = client.Call("Arith.Multiply", args, &reply)
		if err != nil {
			log.Fatal("arith error:", err)
		}
		fmt.Printf("Arith: %d*%d=%d\n", args.A, args.B, reply)
		//异步调用
		//异步方法调用Go 通过 Done channel通知调用结果返回。
		//Go方法是异步的，它返回一个 Call指针对象
		//它的Done是一个channel，如果服务返回，Done就可以得到返回的对象(实际是Call对象，包含Reply和error信息)
		quotient := new(server.Quotient)
		divCall := client.Go("Arith.Divide", args, quotient, nil)
		replyCall := <-divCall.Done // will be equal to divCall
		fmt.Printf("RES: %v\n", quotient)
		fmt.Printf("replyCall: %v\n", replyCall)
		fmt.Printf("divCall: %v\n", divCall)
		// check errors, print, etc.
	}

}
