package bc

import (
	"fmt"
	"sync"
	"testing"
)

func TestBC(t *testing.T) {
	bs := NewBroadcastService()
	chBroadcast := bs.Run()
	chA := bs.Listener()
	chB := bs.Listener()

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

	for i := 0; i < 3; i++ {
		chBroadcast <- i
	}

	bs.RemoveListener(chA)

	for i := 3; i < 6; i++ {
		chBroadcast <- i
	}

	close(chBroadcast)
	wg.Wait()
	t.Log("OK")
}
