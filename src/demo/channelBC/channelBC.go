package main

import (
	"fmt"
	"sync"
	"time"
)

/*
https://science.mroman.ch/gobroadcastchannels.html
*/

func main() {
	if false {
		go Origin()
	}

	if true {
		//go BC_MultiChannel()
		go BC_OneChannel()
	}

	go alive()
	select {}
}

func alive() {
	for {
		time.Sleep(time.Second * 1)
	}
}

/*

The above example will print two lines with the second field being 1 and 2.
Depending on the exact scheduling
it might print A 1 \ B 2, or A 1 \ A 2 or B 1 \ B 2 or any other such variant
but each go routine will print one line as each go routine sees one value.
If we want each value to be seen by all listening go routines we need more channels.
*/
func Origin() {
	ch := make(chan int)

	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		for v := range ch {
			fmt.Println("A", v)
		}
		wg.Done()
	}()

	go func() {
		for v := range ch {
			fmt.Println("B", v)
		}
		wg.Done()
	}()

	ch <- 1
	ch <- 2
	close(ch)
	wg.Wait()
}

/*
Now each listener sees every value.
We're not quite happy with this though.
Now we have to keep track of many channels and what if we want to dynamically add or remove listeners?
Let's look at one first intermediate improvement.
*/
func BC_MultiChannel() {
	chA := make(chan int)
	chB := make(chan int)

	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		for v := range chA {
			fmt.Println("A", v)
		}
		wg.Done()
	}()

	go func() {
		for v := range chB {
			fmt.Println("B", v)
		}
		wg.Done()
	}()

	for i := 0; i < 10; i++ {
		chA <- i
		chB <- i
	}
	close(chA)
	close(chB)
	wg.Wait()
}

/*
Notice the difference? Now we only have to write our values to one single channel chBroadcast.
This is now our broadcast channel.
But now let's get more complicated with dynamically adding and removing listeners!
*/
func BC_OneChannel() {
	chA := make(chan int)
	chB := make(chan int)

	chBroadcast := make(chan int)

	var wg sync.WaitGroup

	wg.Add(3)

	go func() {
		for v := range chBroadcast {
			chA <- v
			chB <- v
		}
		close(chA)
		close(chB)
		wg.Done()
	}()

	go func() {
		for v := range chA {
			fmt.Println("A", v)
		}
		wg.Done()
	}()

	go func() {
		for v := range chB {
			fmt.Println("B", v)
		}
		wg.Done()
	}()

	for i := 0; i < 2; i++ {
		chBroadcast <- i
	}
	close(chBroadcast)
	wg.Wait()
}
