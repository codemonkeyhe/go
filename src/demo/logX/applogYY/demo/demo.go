package main

import "fmt"

//import _ "os"
import "time"
import "demo/applogYY"

var log = applog.MustGetLogger("main")

func logP() {
	for true {
		log.Infof("TTTTTTTT")
		time.Sleep(time.Second)
	}
	//os.Exit(0)
}

/*
只能在windows下编译 linux64的程序
编译没错，无法执行

*/
func main() {
	fmt.Println("hello")
	go logP()
	select {}
}
