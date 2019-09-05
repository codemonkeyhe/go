package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

/*
D:\SvnRepo\go\src\demo\udp>udp -h
Usage of udp:
  -host string
        host
  -port string
        port (default "37")
  -server
        Run server

D:\SvnRepo\go\src\demo\udp>udp -server
Listening on :37
RemoteClient: 0 127.0.0.1:56977

D:\SvnRepo\go\src\demo\udp>udp
Connecting to 127.0.0.1:37
2018-03-22 10:37:11 +0800 CST

*/

func Usage() {
	fmt.Fprint(os.Stderr, "Usage of ", os.Args[0], ":\n")
	flag.PrintDefaults()
	fmt.Fprint(os.Stderr, "\n")
}

var host = flag.String("host", "", "host")
var port = flag.String("port", "37", "port")

func main() {

	bServer := flag.Bool("server", false, "Run server")
	flag.Parse()
	if *bServer {
		addr, err := net.ResolveUDPAddr("udp", *host+":"+*port)
		if err != nil {
			fmt.Println("Can't resolve address: ", err)
			os.Exit(1)
		}

		conn, err := net.ListenUDP("udp", addr)
		if err != nil {
			fmt.Println("Error listening:", err)
			os.Exit(1)
		}
		defer conn.Close()
		fmt.Println("Listening on " + *host + ":" + *port)
		for {
			handleClient(conn)
		}
	} else {
		addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:"+*port)
		if err != nil {
			fmt.Println("Can't resolve address: ", err)
			os.Exit(1)
		}

		conn, err := net.DialUDP("udp", nil, addr)
		if err != nil {
			fmt.Println("Can't dial: ", err)
			os.Exit(1)
		}
		defer conn.Close()
		fmt.Println("Connecting to " + "127.0.0.1:" + *port)
		_, err = conn.Write([]byte(""))
		if err != nil {
			fmt.Println("failed:", err)
			os.Exit(1)
		}

		data := make([]byte, 4)
		_, err = conn.Read(data)
		if err != nil {
			fmt.Println("failed to read UDP msg because of ", err)
			os.Exit(1)
		}
		//时间数是64位的，需要将其转换成一个32位的字节
		t := binary.BigEndian.Uint32(data)
		fmt.Println(time.Unix(int64(t), 0).String())

		os.Exit(0)

	}

}

func handleClient(conn *net.UDPConn) {
	data := make([]byte, 1024)
	n, remoteAddr, err := conn.ReadFromUDP(data)
	if err != nil {
		fmt.Println("failed to read UDP msg because of ", err.Error())
		return
	}

	daytime := time.Now().Unix()
	fmt.Println("RemoteClient:", n, remoteAddr)
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(daytime))
	conn.WriteToUDP(b, remoteAddr)
}
