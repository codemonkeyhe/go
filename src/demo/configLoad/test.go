// goroutine project goroutine.go
package main

import (
	"fmt"
	"log/syslog"
	"runtime"
	"sync"
	//	"common/second/applog"
	//	"encoding/json"
	//"strings"
	//"net/url"
	//	"net"
	//	"os"
	//	"io/ioutil"
	//	"net/http"

	"time"
)

var _ = runtime.GOMAXPROCS(1)

/******Output*****

hello world

****************/

func testR(j int) {
	for {
		fmt.Println("TEST")
		GetInstance().readC(int8(j))
		time.Sleep(time.Second * 2)
	}
}
