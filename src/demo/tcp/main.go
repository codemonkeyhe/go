package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"sync"
)

func Usage() {
	fmt.Fprint(os.Stderr, "Usage of ", os.Args[0], ":\n")
	flag.PrintDefaults()
	fmt.Fprint(os.Stderr, "\n")
}

var host = flag.String("host", "", "host")
var port = flag.String("port", "3333", "port")
var bOrigin = true

/*
D:\SvnRepo\go\src\demo\tcp>tcp --help
Usage of tcp:
  -host string
        host
  -port string
        port (default "3333")
  -server
        Run server

D:\SvnRepo\go\src\demo\tcp>tcp -server
Listening on :3333
Received message 127.0.0.1:21331 -> 127.0.0.1:3333
Receive REQ!


D:\SvnRepo\go\src\demo\tcp>tcp
Connecting to 127.0.0.1:3333
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

func main() {
	flag.Usage = Usage
	bServer := flag.Bool("server", false, "Run server")
	flag.Parse()
	if *bServer {

		var l net.Listener
		var err error
		l, err = net.Listen("tcp", *host+":"+*port)
		if err != nil {
			fmt.Println("Error listening:", err)
			os.Exit(1)
		}
		defer l.Close()
		fmt.Println("Listening on " + *host + ":" + *port)

		for {
			conn, err := l.Accept()
			if err != nil {
				fmt.Println("Error accepting: ", err)
				os.Exit(1)
			}
			//logs an incoming message
			fmt.Printf("Received message %s -> %s \n", conn.RemoteAddr(), conn.LocalAddr())

			// Handle connections in a new goroutine.
			go handleRequest(conn)
		}

	} else {

		h := "127.0.0.1"
		conn, err := net.Dial("tcp", h+":"+*port)
		if err != nil {
			fmt.Println("Error connecting:", err)
			os.Exit(1)
		}
		defer conn.Close()

		fmt.Println("Connecting to " + h + ":" + *port)

		if bOrigin {
			done := make(chan string)
			go handleWrite(conn, done)
			go handleRead(conn, done)

			fmt.Println(<-done)
			fmt.Println(<-done)
		} else {
			var wg sync.WaitGroup
			wg.Add(2)

			go handleWriteWg(conn, &wg)
			go handleReadWg(conn, &wg)

			wg.Wait()
		}

	}

}

func handleRequest(conn net.Conn) {
	defer conn.Close()
	fmt.Println("Receive REQ!")
	for {
		io.Copy(conn, conn)
	}
	//不会执行
	fmt.Println("ERROR!NO supposed HERE!")
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

func handleWriteWg(conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	for i := 10; i > 0; i-- {
		_, e := conn.Write([]byte("hello " + strconv.Itoa(i) + "\r\n"))

		if e != nil {
			fmt.Println("Error to send message because of ", e.Error())
			break
		}
	}
	fmt.Println("Send REQ!")
}

func handleReadWg(conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	reader := bufio.NewReader(conn)
	for i := 1; i <= 10; i++ {
		line, err := reader.ReadString(byte('\n'))
		if err != nil {
			fmt.Print("Error to read message because of ", err)
			return
		}
		fmt.Print(line)
	}
	fmt.Println("Receive RESP!")
}
