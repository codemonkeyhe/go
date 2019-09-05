package main

// "github.com/robteix/testmod"
// https://roberto.selbach.ca/tag/modules/
// https://roberto.selbach.ca/intro-to-go-modules/
//
// "github.com/robteix/testmod"
// "local.com/xyz/testmod"

import (
	"fmt"
	"time"

	"demo.com/gomod_demo/service/dao"

	"local.com/xyz/testmod"
)

func main() {
	dao.DaoInit()
	fmt.Println(testmod.Hi("roberto"))
	time.Sleep(time.Second * 3)
}
