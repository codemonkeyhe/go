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

	"local.com/xyz/testmod"
)

func main() {
	fmt.Println(testmod.Hi("roberto"))
	time.Sleep(time.Second * 3)
}
