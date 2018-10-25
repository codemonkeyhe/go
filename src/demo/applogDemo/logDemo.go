package main

import "fmt"

//import _ "os"
import "time"
import "common/applog"

func logP() {
	for true {
		applog.Info("TTTTTTTT")
		time.Sleep(time.Second)
	}
	//os.Exit(0)
}

func main() {
	go logP()
	fmt.Println("hello")
	select {}
}
