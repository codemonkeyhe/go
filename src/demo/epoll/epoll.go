package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"syscall"
)

/*
syscall包是如何实现的
https://groups.google.com/forum/#!topic/golang-china/dVOsyjAOegQ


http://colobu.com/2017/11/29/event-loop-networking-in-Go/
对于网络编程，Go标准库和运行时内部采用 epoll/kqueue/IoCompletionPort来实现基于 event-loop的网络异步处理，
但是通过netpoll的方式对外提供同步的访问。
具体代码可以参考 runtime/netpoll、net和internal/poll。


不过 net 包封装的同步 api 更好用（底层使用 epoll 实现）

[monkey@bogon epoll]$./epoll -server
Listen at addr: {Port:2000 Addr:[0 0 0 0] raw:{Family:2 Port:53255 Addr:[0 0 0 0] Zero:[0 0 0 0 0 0 0 0]}}
new connFd acceptd!connFd: 6
receive data on connFD: 6
receive data on connFD: 6
receive data on connFD: 6
receive data on connFD: 6
>>> hello 10
hello 9
hello 8
hello 7
hello 6
hello 5
hello 4
hello 3
hello 2
hello 1
receive data on connFD: 6
<<< hello 10
hello 9
hello 8
hello 7
hello 6
hello 5
hello 4
hello 3
hello 2
hello 1
^C

[monkey@bogon epoll]$./epoll
Connecting to 127.0.0.1:2000
Send REQ!
Sent
hello 10
hello 9
hello 8
hello 7
hello 6
hello 5
hello 4
hello 3
hello 2
hello 1
Receive RESP!
Read



*/

func Usage() {
	fmt.Fprint(os.Stderr, "Usage of ", os.Args[0], ":\n")
	flag.PrintDefaults()
	fmt.Fprint(os.Stderr, "\n")
}

const (
	EPOLLET        = 1 << 31
	MaxEpollEvents = 32
)

func echo(fd int) {
	defer syscall.Close(fd)
	var buf [32 * 1024]byte
	for {
		nbytes, e := syscall.Read(fd, buf[:])
		if nbytes > 0 {
			fmt.Printf(">>> %s", buf)
			syscall.Write(fd, buf[:nbytes])
			fmt.Printf("<<< %s", buf)
		}
		if e != nil {
			break
		}
	}
}

func main() {
	flag.Usage = Usage
	bServer := flag.Bool("server", false, "Run server")
	flag.Parse()
	if *bServer {
		runServer()
	} else {

		runClient()
	}

}

var port int = 2000

func runServer() {
	var event syscall.EpollEvent
	var events [MaxEpollEvents]syscall.EpollEvent

	fd, err := syscall.Socket(syscall.AF_INET, syscall.O_NONBLOCK|syscall.SOCK_STREAM, 0)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer syscall.Close(fd)

	if err = syscall.SetNonblock(fd, true); err != nil {
		fmt.Println("setnonblock1: ", err)
		os.Exit(1)
	}

	addr := syscall.SockaddrInet4{Port: port}
	copy(addr.Addr[:], net.ParseIP("0.0.0.0").To4())

	syscall.Bind(fd, &addr)
	syscall.Listen(fd, 10)
	fmt.Printf("Listen at addr: %+v\n", addr)

	epfd, e := syscall.EpollCreate1(0)
	if e != nil {
		fmt.Println("epoll_create1: ", e)
		os.Exit(1)
	}
	defer syscall.Close(epfd)

	event.Events = syscall.EPOLLIN
	event.Fd = int32(fd)
	if e = syscall.EpollCtl(epfd, syscall.EPOLL_CTL_ADD, fd, &event); e != nil {
		fmt.Println("epoll_ctl: ", e)
		os.Exit(1)
	}

	for {
		nevents, e := syscall.EpollWait(epfd, events[:], -1)
		if e != nil {
			fmt.Println("epoll_wait: ", e)
			break
		}

		for ev := 0; ev < nevents; ev++ {
			if int(events[ev].Fd) == fd {
				//如果是监听的fd有IO，说明有新的链接来了
				connFd, _, err := syscall.Accept(fd)
				if err != nil {
					fmt.Println("accept: ", err)
					continue
				}
				fmt.Printf("new connFd acceptd!connFd: %+v\n", connFd)
				syscall.SetNonblock(fd, true)
				//新链接connFd设置为边缘触发
				event.Events = syscall.EPOLLIN | EPOLLET
				event.Fd = int32(connFd)
				if err := syscall.EpollCtl(epfd, syscall.EPOLL_CTL_ADD, connFd, &event); err != nil {
					fmt.Print("epoll_ctl: ", connFd, err)
					os.Exit(1)
				}
			} else {
				//非监听fd，说明是已accept的连接
				fmt.Printf("receive data on connFD: %+v\n", events[ev].Fd)
				go echo(int(events[ev].Fd))
			}
		}

	}
}

func runClient() {
	h := "127.0.0.1"
	port := "2000"
	conn, err := net.Dial("tcp", h+":"+port)
	if err != nil {
		fmt.Println("Error connecting:", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("Connecting to " + h + ":" + port)

	done := make(chan string)
	go handleWrite(conn, done)
	go handleRead(conn, done)

	fmt.Println(<-done)
	fmt.Println(<-done)
}

func handleWrite(conn net.Conn, done chan string) {
	for i := 10; i > 0; i-- {
		_, e := conn.Write([]byte("hello " + strconv.Itoa(i) + "\r\n"))

		if e != nil {
			fmt.Println("Error to send message because of ", e.Error())
			break
		}
	}
	fmt.Println("Send REQ!")
	done <- "Sent"
}

func handleRead(conn net.Conn, done chan string) {
	buf := make([]byte, 1024)
	reqLen, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error to read message because of ", err)
		return
	}
	fmt.Println(string(buf[:reqLen-1]))
	fmt.Println("Receive RESP!")
	done <- "Read"
}
