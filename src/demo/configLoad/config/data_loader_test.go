package config

import (
	"strconv"
	"testing"
	"time"
)

func Test_loader(t *testing.T) {
	dataL := GetInstance()
	dataL.Init()
	dataL.Stat()
	dataL.Start()

	//并发写
	go func() {
		for {
			for i := 0; i < 10000; i++ {
				val := strconv.FormatInt(int64(i+20000), 10)
				go dataL.writeC(val, val)
			}
			time.Sleep(time.Millisecond * 1000)
		}
	}()
	//并发读
	go func() {
		for {
			for i := 0; i < 10000; i++ {
				val := strconv.FormatInt(int64(i+20000), 10)
				go dataL.readC(val)
			}
			time.Sleep(time.Millisecond * 3000)
		}
	}()

}
