// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build go1.7

package rate

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLimit(t *testing.T) {
	if Limit(10) == Inf {
		t.Errorf("Limit(10) == Inf should be false")
	}
}

func closeEnough(a, b Limit) bool {
	return (math.Abs(float64(a)/float64(b)) - 1.0) < 1e-9
}

func TestEvery(t *testing.T) {
	cases := []struct {
		interval time.Duration
		lim      Limit
	}{
		{0, Inf},
		{-1, Inf},
		{1 * time.Nanosecond, Limit(1e9)},
		{1 * time.Microsecond, Limit(1e6)},
		{1 * time.Millisecond, Limit(1e3)},
		{10 * time.Millisecond, Limit(100)},
		{100 * time.Millisecond, Limit(10)},
		{1 * time.Second, Limit(1)},
		{2 * time.Second, Limit(0.5)},
		{time.Duration(2.5 * float64(time.Second)), Limit(0.4)},
		{4 * time.Second, Limit(0.25)},
		{10 * time.Second, Limit(0.1)},
		{time.Duration(math.MaxInt64), Limit(1e9 / float64(math.MaxInt64))},
	}
	for _, tc := range cases {
		lim := Every(tc.interval)
		if !closeEnough(lim, tc.lim) {
			t.Errorf("Every(%v) = %v want %v", tc.interval, lim, tc.lim)
		}
	}
}

const (
	d = 100 * time.Millisecond
)

var (
	t0    = time.Now()
	t1    = t0.Add(time.Duration(1) * d)
	t2    = t0.Add(time.Duration(2) * d)
	t3    = t0.Add(time.Duration(3) * d)
	t4    = t0.Add(time.Duration(4) * d)
	t5    = t0.Add(time.Duration(5) * d)
	t6    = t0.Add(time.Duration(6) * d)
	t9    = t0.Add(time.Duration(9) * d)
	t999  = t0.Add(time.Duration(999) * d)
	t1000 = t0.Add(time.Duration(1000) * d)
	t1002 = t0.Add(time.Duration(1002) * d)
	t2000 = t0.Add(time.Duration(2000) * d)
	t2001 = t0.Add(time.Duration(2001) * d)
	t19   = t0.Add(time.Duration(19) * d)
	t20   = t0.Add(time.Duration(20) * d)
	t40   = t0.Add(time.Duration(40) * d)
	t60   = t0.Add(time.Duration(60) * d)
	t80   = t0.Add(time.Duration(80) * d)
	t100  = t0.Add(time.Duration(100) * d)
	t81   = t0.Add(time.Duration(81) * d)
	t36   = t0.Add(time.Duration(36) * d)
)

type allow struct {
	t  time.Time
	n  int
	ok bool
}

func run(t *testing.T, lim *Limiter, allows []allow) {
	for i, allow := range allows {
		ok := lim.AllowN(allow.t, allow.n)
		if ok != allow.ok {
			t.Errorf("step %d: lim.AllowN(%v, %v) = %v want %v",
				i, allow.t, allow.n, ok, allow.ok)
		}
	}
}

func TestLimiterBurst1(t *testing.T) {
	run(t, NewLimiter(10, 1), []allow{
		{t0, 1, true},
		{t0, 1, false},
		{t0, 1, false},
		{t1, 1, true},
		{t1, 1, false},
		{t1, 1, false},
		{t2, 2, false}, // burst size is 1, so n=2 always fails
		{t2, 1, true},
		{t2, 1, false},
	})
}

func TestLimiterBurst3(t *testing.T) {
	run(t, NewLimiter(10, 3), []allow{
		{t0, 2, true},
		{t0, 2, false},
		{t0, 1, true},
		{t0, 1, false},
		{t1, 4, false},
		{t2, 1, true},
		{t3, 1, true},
		{t4, 1, true},
		{t4, 1, true},
		{t4, 1, false},
		{t4, 1, false},
		{t9, 3, true},
		{t9, 0, true},
	})
}

func TestLimiterJumpBackwards(t *testing.T) {
	run(t, NewLimiter(10, 3), []allow{
		{t1, 1, true}, // start at t1
		{t0, 1, true}, // jump back to t0, two tokens remain
		{t0, 1, true},
		{t0, 1, false},
		{t0, 1, false},
		{t1, 1, true}, // got a token
		{t1, 1, false},
		{t1, 1, false},
		{t2, 1, true}, // got another token
		{t2, 1, false},
		{t2, 1, false},
	})
}

// Ensure that tokensFromDuration doesn't produce
// rounding errors by truncating nanoseconds.
// See golang.org/issues/34861.
func TestLimiter_noTruncationErrors(t *testing.T) {
	if !NewLimiter(0.7692307692307693, 1).Allow() {
		t.Fatal("expected true")
	}
}

func TestSimultaneousRequests(t *testing.T) {
	const (
		limit       = 1
		burst       = 5
		numRequests = 15
	)
	var (
		wg    sync.WaitGroup
		numOK = uint32(0)
	)

	// Very slow replenishing bucket.
	lim := NewLimiter(limit, burst)

	// Tries to take a token, atomically updates the counter and decreases the wait
	// group counter.
	f := func() {
		defer wg.Done()
		if ok := lim.Allow(); ok {
			atomic.AddUint32(&numOK, 1)
		}
	}

	wg.Add(numRequests)
	for i := 0; i < numRequests; i++ {
		go f()
	}
	wg.Wait()
	if numOK != burst {
		t.Errorf("numOK = %d, want %d", numOK, burst)
	}
}

func TestLongRunningQPS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	if runtime.GOOS == "openbsd" {
		t.Skip("low resolution time.Sleep invalidates test (golang.org/issue/14183)")
		return
	}

	// The test runs for a few seconds executing many requests and then checks
	// that overall number of requests is reasonable.
	const (
		limit = 100
		burst = 100
	)
	var numOK = int32(0)

	lim := NewLimiter(limit, burst)

	var wg sync.WaitGroup
	f := func() {
		if ok := lim.Allow(); ok {
			atomic.AddInt32(&numOK, 1)
		}
		wg.Done()
	}

	start := time.Now()
	end := start.Add(5 * time.Second)
	for time.Now().Before(end) {
		wg.Add(1)
		go f()

		// This will still offer ~500 requests per second, but won't consume
		// outrageous amount of CPU.
		time.Sleep(2 * time.Millisecond)
	}
	wg.Wait()
	elapsed := time.Since(start)
	ideal := burst + (limit * float64(elapsed) / float64(time.Second))

	// We should never get more requests than allowed.
	if want := int32(ideal + 1); numOK > want {
		t.Errorf("numOK = %d, want %d (ideal %f)", numOK, want, ideal)
	}
	// We should get very close to the number of requests allowed.
	if want := int32(0.999 * ideal); numOK < want {
		t.Errorf("numOK = %d, want %d (ideal %f)", numOK, want, ideal)
	}
}

type request struct {
	t   time.Time
	n   int
	act time.Time
	ok  bool
}

// dFromDuration converts a duration to a multiple of the global constant d
func dFromDuration(dur time.Duration) int {
	// Adding a millisecond to be swallowed by the integer division
	// because we don't care about small inaccuracies
	return int((dur + time.Millisecond) / d)
}

// dSince returns multiples of d since t0
func dSince(t time.Time) int {
	return dFromDuration(t.Sub(t0))
}

func runReserve(t *testing.T, lim *Limiter, req request) *Reservation {
	return runReserveMax(t, lim, req, InfDuration)
}

func runReserveMax(t *testing.T, lim *Limiter, req request, maxReserve time.Duration) *Reservation {
	r := lim.reserveN(req.t, req.n, maxReserve)
	if r.ok && (dSince(r.timeToAct) != dSince(req.act)) || r.ok != req.ok {
		t.Errorf("lim.reserveN(t%d, %v, %v) = (t%d, %v) want (t%d, %v)",
			dSince(req.t), req.n, maxReserve, dSince(r.timeToAct), r.ok, dSince(req.act), req.ok)
	}
	return &r
}

func TestSimpleReserve(t *testing.T) {
	lim := NewLimiter(10, 2)

	runReserve(t, lim, request{t0, 2, t0, true})
	runReserve(t, lim, request{t0, 2, t2, true})
	runReserve(t, lim, request{t3, 2, t4, true})
}

func TestZeroBurst(t *testing.T) {
	lim := NewLimiter(10, 0)

	runReserve(t, lim, request{t0, 2, t1000, false})
	runReserve(t, lim, request{t0, 2, t1000, false})
	runReserve(t, lim, request{t3, 2, t1000, false})
}

/*
func TestZeroLimit(t *testing.T) {
	lim := NewLimiter(0, 10)
	runReserve(t, lim, request{t0, 2, t0, true})
	r := runReserve(t, lim, request{t0, 8, t0, true})
	fmt.Printf("curToken=%f\n", r.lim.tokens) // 0
	runReserve(t, lim, request{t1, 2, t1, true})
}
*/

func TestMix(t *testing.T) {
	lim := NewLimiter(10, 2)

	runReserve(t, lim, request{t0, 3, t1, false}) // should return false because n > Burst
	runReserve(t, lim, request{t0, 2, t0, true})
	run(t, lim, []allow{{t1, 2, false}}) // not enought tokens - don't allow
	runReserve(t, lim, request{t1, 2, t2, true})
	run(t, lim, []allow{{t1, 1, false}}) // negative tokens - don't allow
	run(t, lim, []allow{{t3, 1, true}})
}

func TestCancelInvalid(t *testing.T) {
	lim := NewLimiter(10, 2)

	runReserve(t, lim, request{t0, 2, t0, true})
	r := runReserve(t, lim, request{t0, 3, t3, false})
	r.CancelAt(t0)                               // should have no effect
	runReserve(t, lim, request{t0, 2, t2, true}) // did not get extra tokens
}

func TestCancelLast(t *testing.T) {
	lim := NewLimiter(10, 2)

	r0 := runReserve(t, lim, request{t0, 2, t0, true})
	fmt.Printf("curToken=%f\n", r0.lim.tokens)
	r := runReserve(t, lim, request{t0, 2, t2, true})
	fmt.Printf("curToken=%f\n", r.lim.tokens)
	r.CancelAt(t1) // got 2 tokens back
	fmt.Printf("After Cancel, curToken=%f\n", r.lim.tokens)
	runReserve(t, lim, request{t1, 2, t2, true})
}

func TestCancelTooLate(t *testing.T) {
	lim := NewLimiter(10, 2)

	r0 := runReserve(t, lim, request{t0, 2, t0, true})
	fmt.Printf("curToken=%f\n", r0.lim.tokens) //0
	r := runReserve(t, lim, request{t0, 2, t2, true})
	fmt.Printf("curToken=%f\n", r.lim.tokens) //-2
	r.CancelAt(t3)                            // too late to cancel - should have no effect
	fmt.Printf("curToken=%f\n", r.lim.tokens) //-2
	runReserve(t, lim, request{t3, 2, t4, true})
	fmt.Printf("curToken=%f\n", r.lim.tokens) //-1
}

func TestCancel0Tokens(t *testing.T) {
	lim := NewLimiter(10, 2)

	runReserve(t, lim, request{t0, 2, t0, true})       // 0
	r := runReserve(t, lim, request{t0, 1, t1, true})  //-1
	r1 := runReserve(t, lim, request{t0, 1, t2, true}) //-2
	fmt.Printf("curToken=%f\n", r1.lim.tokens)         //-2
	// why is not 1 token back
	r.CancelAt(t0)                                         // got 0 tokens back
	fmt.Printf("after Cancel curToken=%f\n", r.lim.tokens) //-2 , NOT -1
	runReserve(t, lim, request{t0, 1, t3, true})
}

func TestCancel1Token(t *testing.T) {
	lim := NewLimiter(10, 2)

	runReserve(t, lim, request{t0, 2, t0, true})           // 0
	r := runReserve(t, lim, request{t0, 2, t2, true})      // -2
	runReserve(t, lim, request{t0, 1, t3, true})           // -3
	fmt.Printf("curToken=%f\n", r.lim.tokens)              // -3
	r.CancelAt(t2)                                         // got 1 token back
	fmt.Printf("after Cancel curToken=%f\n", r.lim.tokens) // 0  NOT 1
	runReserve(t, lim, request{t2, 2, t4, true})
}

func TestCancelTokenM(t *testing.T) {
	lim := NewLimiter(10, 1000)

	runReserve(t, lim, request{t0, 1000, t0, true})   // 0
	r := runReserve(t, lim, request{t0, 2, t2, true}) // -2

	runReserve(t, lim, request{t0, 1000, t1002, true}) // -1002
	fmt.Printf("curToken=%f\n", r.lim.tokens)          // -3
	r.CancelAt(t2)                                     // restore=-998  <=0  got 0 back
	fmt.Printf("after Cancel curToken=%f\n", r.lim.tokens)
	//runReserve(t, lim, request{t2, 2, t4, true})
}

func TestCancelTokenN(t *testing.T) {
	lim := NewLimiter(10, 1000)

	runReserve(t, lim, request{t0, 1000, t0, true})         // 0
	r := runReserve(t, lim, request{t0, 1000, t1000, true}) // -1000

	r1 := runReserve(t, lim, request{t0, 2, t1002, true}) // -1002
	fmt.Printf("curToken=%f\n", r.lim.tokens)             // -1002
	if false {                                            //return at t2
		r.CancelAt(t2) //return 998 = (1000  - (t1002-t1000))
		// -1002 + 2 +998=  -2
		fmt.Printf("after Cancel curToken=%f\n", r.lim.tokens) //-2
		runReserve(t, lim, request{t2, 1, t5, true})
	}
	if false { //return at t5
		r.CancelAt(t5) //return 998 = (1000  - (t1002-t1000))
		// -1002 + 5 +998=  1
		fmt.Printf("after Cancel curToken=%f\n", r.lim.tokens) // 1
		runReserve(t, lim, request{t5, 2, t6, true})
	}

	if true { // return at t999
		r.CancelAt(t999) //return 998 = (1000  - (t1002-t1000))
		// -1002 + 999 +998= 995
		fmt.Printf("after Cancel curToken=%f\n", r.lim.tokens)            // 995
		runReserve(t, lim, request{t1002, 998, t1002, true})              // 995+3-998=0
		fmt.Printf("curToken=%f\n", r.lim.tokens)                         // 0
		fmt.Printf("lastTime=%v\n", r.lim.last)                           // 0
		fmt.Printf("r1.Act  =%v r1.tokens=%d\n", r1.timeToAct, r1.tokens) // 0
	}

	// for  TestCancelTokenN , simulate QPS > burst
	if false { // return at t999
		r.CancelAt(t999) //return 1000
		// -1002 + 999 +1000 = 997
		fmt.Printf("after Cancel curToken=%f\n", r.lim.tokens)            // 997
		r3 := runReserve(t, lim, request{t1002, 1000, t1002, true})       // 997+3-1000=0
		fmt.Printf("curToken=%f\n", r.lim.tokens)                         // 0
		fmt.Printf("lastTime=%v\n", r.lim.last)                           // 0
		fmt.Printf("r1.Act  =%v r1.tokens=%d\n", r1.timeToAct, r1.tokens) // 1000
		fmt.Printf("r3.Act  =%v r3.tokens=%d\n", r3.timeToAct, r3.tokens) // 2
		// 可以看到，同一时间点，r1和r3都执行了，最终瞬时QPS=1000+2=1002>1000
	}

}

//不同的归还顺序对lim.Tokens有影响
func TestCancelToken10(t *testing.T) {
	lim := NewLimiter(10, 20)

	r0 := runReserve(t, lim, request{t0, 20, t0, true})   // 0
	r1 := runReserve(t, lim, request{t0, 20, t20, true})  // -20
	r2 := runReserve(t, lim, request{t0, 20, t40, true})  // -40
	r3 := runReserve(t, lim, request{t0, 20, t60, true})  // -60
	r4 := runReserve(t, lim, request{t0, 20, t80, true})  // -80
	r5 := runReserve(t, lim, request{t0, 20, t100, true}) // -100
	fmt.Printf("curToken=%f\n", r0.lim.tokens)            // -100

	//order1 -61
	// 实质上只有r5取消成功了
	if false {
		//return  20-(t100-t20)=20-80=-60 <=0
		r1.CancelAt(t19)
		fmt.Printf("after Cancel curToken=%f\n", r0.lim.tokens) // -100  last=t0 not t19
		//return  20-(t100-t40)=20-60=-40 <=0
		r2.CancelAt(t19)
		fmt.Printf("after Cancel curToken=%f\n", r0.lim.tokens)
		r3.CancelAt(t19)
		fmt.Printf("after Cancel curToken=%f\n", r0.lim.tokens)
		r4.CancelAt(t19)                                        // 20- 20 = 0 <=0
		fmt.Printf("after Cancel curToken=%f\n", r0.lim.tokens) // -100 ok=false
		r5.CancelAt(t19)                                        // 20- 0 = 20
		fmt.Printf("after Cancel curToken=%f\n", r0.lim.tokens) // -100+19+20 = -61
		runReserve(t, lim, request{t20, 1, t81, true})          // -61 + 1 -1 = -61, t20到t81才能提供61个token
		fmt.Printf("curToken=%f\n", r0.lim.tokens)              // -61, now=t20
	}
	//order2 0
	// 实质上 r5~r1都取消成功了
	if true {
		// restoreTokens= 20-(t100-t100) = 20
		r5.CancelAt(t19)                                        // -100 +19 +20 = -61
		fmt.Printf("after Cancel curToken=%f\n", r0.lim.tokens) // -61 , lastEvent=t80

		// restoreTokens= 20-(t80-t80) = 20
		r4.CancelAt(t19)                                        // -61+20 =-41
		fmt.Printf("after Cancel curToken=%f\n", r0.lim.tokens) // -41

		r3.CancelAt(t19)
		fmt.Printf("after Cancel curToken=%f\n", r0.lim.tokens) // -41 +20 = -21

		r2.CancelAt(t19)
		fmt.Printf("after Cancel curToken=%f\n", r0.lim.tokens) // -21 +20 = -1, lastEvent=t20
		fmt.Printf("lastEvent=%v\n      t20=%v\n", r0.lim.lastEvent, t20)

		// restoreTokens= 20-(t20-t20) = 20
		r1.CancelAt(t19)                                        // -1+20 = 19
		fmt.Printf("after Cancel curToken=%f\n", r0.lim.tokens) // 19

		runReserve(t, lim, request{t20, 20, t20, true}) // 19+1 -20 = 0
		fmt.Printf("curToken=%f\n", r0.lim.tokens)      // 0
	}

}

func TestCancelTokenX(t *testing.T) {
	lim := NewLimiter(10, 1000)

	runReserve(t, lim, request{t0, 1000, t0, true})         // 0
	r := runReserve(t, lim, request{t0, 1000, t1000, true}) // -1000
	runReserve(t, lim, request{t0, 1000, t2000, true})      // -2000
	fmt.Printf("curToken=%f\n", r.lim.tokens)               // -2000
	r.CancelAt(t2)                                          // restore=0  <=0  got 0 back
	fmt.Printf("after Cancel curToken=%f\n", r.lim.tokens)  // -2000
	runReserve(t, lim, request{t2, 1, t2001, true})
}

func TestCancelMulti(t *testing.T) {
	lim := NewLimiter(10, 4)

	runReserve(t, lim, request{t0, 4, t0, true})            // 0
	rA := runReserve(t, lim, request{t0, 3, t3, true})      //t0. lim.tokens=-3
	runReserve(t, lim, request{t0, 1, t4, true})            //t0. lim.tokens=-4
	rC := runReserve(t, lim, request{t0, 1, t5, true})      //t0. lim.tokens=-5
	rC.CancelAt(t1)                                         // get 1 token back
	fmt.Printf("after Cancel curToken=%f\n", rA.lim.tokens) // lim.tokens=-5+1+1 = -3  last=t1
	// get 2 tokens back, as if C was never reserved
	//restoreTokens=3-(lastEvent-t3) = 3-(t4-t3) = 2
	rA.CancelAt(t1)
	fmt.Printf("after Cancel curToken=%f\n", rA.lim.tokens) // lim.tokens=-3+2=-1
	runReserve(t, lim, request{t1, 3, t5, true})
}

func TestReserveJumpBack(t *testing.T) {
	lim := NewLimiter(10, 2)

	runReserve(t, lim, request{t1, 2, t1, true}) // start at t1
	runReserve(t, lim, request{t0, 1, t1, true}) // should violate Limit,Burst
	runReserve(t, lim, request{t2, 2, t3, true})
}

func TestReserveJumpBackCancel(t *testing.T) {
	lim := NewLimiter(10, 2)

	runReserve(t, lim, request{t1, 2, t1, true}) // start at t1
	r := runReserve(t, lim, request{t1, 2, t3, true})
	runReserve(t, lim, request{t1, 1, t4, true})
	r.CancelAt(t0)                               // cancel at t0, get 1 token back
	runReserve(t, lim, request{t1, 2, t4, true}) // should violate Limit,Burst
}

func TestReserveSetLimit(t *testing.T) {
	lim := NewLimiter(5, 2)

	runReserve(t, lim, request{t0, 2, t0, true})
	runReserve(t, lim, request{t0, 2, t4, true})
	lim.SetLimitAt(t2, 10)
	runReserve(t, lim, request{t2, 1, t4, true}) // violates Limit and Burst
}

func TestReserveSetBurst(t *testing.T) {
	lim := NewLimiter(5, 2)

	runReserve(t, lim, request{t0, 2, t0, true})
	runReserve(t, lim, request{t0, 2, t4, true})
	lim.SetBurstAt(t3, 4)
	runReserve(t, lim, request{t0, 4, t9, true}) // violates Limit and Burst
}

func TestReserveSetLimitCancel(t *testing.T) {
	lim := NewLimiter(5, 2)

	runReserve(t, lim, request{t0, 2, t0, true})
	r := runReserve(t, lim, request{t0, 2, t4, true})
	lim.SetLimitAt(t2, 10)
	r.CancelAt(t2) // 2 tokens back
	runReserve(t, lim, request{t2, 2, t3, true})
}

func TestReserveMax(t *testing.T) {
	lim := NewLimiter(10, 2)
	maxT := d

	runReserveMax(t, lim, request{t0, 2, t0, true}, maxT)
	runReserveMax(t, lim, request{t0, 1, t1, true}, maxT)  // reserve for close future
	runReserveMax(t, lim, request{t0, 1, t2, false}, maxT) // time to act too far in the future
}

type wait struct {
	name   string
	ctx    context.Context
	n      int
	delay  int // in multiples of d
	nilErr bool
}

func runWait(t *testing.T, lim *Limiter, w wait) {
	start := time.Now()
	err := lim.WaitN(w.ctx, w.n)
	delay := time.Now().Sub(start)
	if (w.nilErr && err != nil) || (!w.nilErr && err == nil) || w.delay != dFromDuration(delay) {
		errString := "<nil>"
		if !w.nilErr {
			errString = "<non-nil error>"
		}
		t.Errorf("lim.WaitN(%v, lim, %v) = %v with delay %v ; want %v with delay %v",
			w.name, w.n, err, delay, errString, d*time.Duration(w.delay))
	}
}

func TestWaitSimple(t *testing.T) {
	lim := NewLimiter(10, 3)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	runWait(t, lim, wait{"already-cancelled", ctx, 1, 0, false})

	runWait(t, lim, wait{"exceed-burst-error", context.Background(), 4, 0, false})

	runWait(t, lim, wait{"act-now", context.Background(), 2, 0, true})
	runWait(t, lim, wait{"act-later", context.Background(), 3, 2, true})
}

func TestWaitCancel(t *testing.T) {
	lim := NewLimiter(10, 3)

	ctx, cancel := context.WithCancel(context.Background())
	runWait(t, lim, wait{"act-now", ctx, 2, 0, true}) // after this lim.tokens = 1
	go func() {
		time.Sleep(d)
		cancel()
	}()
	runWait(t, lim, wait{"will-cancel", ctx, 3, 1, false})
	// should get 3 tokens back, and have lim.tokens = 2
	t.Logf("tokens:%v last:%v lastEvent:%v", lim.tokens, lim.last, lim.lastEvent)
	runWait(t, lim, wait{"act-now-after-cancel", context.Background(), 2, 0, true})
}

func TestWaitTimeout(t *testing.T) {
	lim := NewLimiter(10, 3)

	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()
	runWait(t, lim, wait{"act-now", ctx, 2, 0, true})
	runWait(t, lim, wait{"w-timeout-err", ctx, 3, 0, false})
}

func TestWaitInf(t *testing.T) {
	lim := NewLimiter(Inf, 0)

	runWait(t, lim, wait{"exceed-burst-no-error", context.Background(), 3, 0, true})
}

func BenchmarkAllowN(b *testing.B) {
	lim := NewLimiter(Every(1*time.Second), 1)
	now := time.Now()
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lim.AllowN(now, 1)
		}
	})
}

func BenchmarkWaitNNoDelay(b *testing.B) {
	lim := NewLimiter(Limit(b.N), b.N)
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lim.WaitN(ctx, 1)
	}
}
