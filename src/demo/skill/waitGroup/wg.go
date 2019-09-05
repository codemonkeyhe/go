package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func process(id int) {
	fmt.Println(id)
}

func write_db(id int) {
	fmt.Printf("Write:%d\n", id)
}

func op1(ctx context.Context, res chan<- int) int {
	time.Sleep(6 * time.Second)
	res <- 100
	return 100
}

func op2(ctx context.Context, res chan<- int) int {
	time.Sleep(2 * time.Second)
	res <- 200
	return 200
}

func op3(ctx context.Context, res chan<- int) int {
	time.Sleep(2 * time.Second)
	res <- 300
	return 300
}

func opp1(ctx context.Context) int {
	time.Sleep(6 * time.Second)
	return 100
}

func opp2(ctx context.Context) int {
	time.Sleep(2 * time.Second)
	return 200
}

func opp3(ctx context.Context) int {
	time.Sleep(2 * time.Second)
	return 300
}

type Element struct {
	JobType int
	JobRes  int
}

func doWork(ctx context.Context, pe *Element, wg *sync.WaitGroup) {
	defer wg.Done()

	done := make(chan struct{})
	go func() {
		// do some work on element
		fmt.Printf("Begin: %+v\n", *pe)
		if pe.JobType == 1 {
			pe.JobRes = opp1(ctx)
		} else if pe.JobType == 2 {
			pe.JobRes = opp2(ctx)
		} else if pe.JobType == 3 {
			pe.JobRes = opp3(ctx)
		}
		fmt.Printf("End: %+v\n", *pe)
		done <- struct{}{} // signal work is done
	}()

	select {
	case <-done:
		{
			// work completed in time
		}
	case <-ctx.Done():
		{
			fmt.Println("Timeout: Err: ", ctx.Err(), " Ele=", *pe)
			// timeout reached
		}
	}
}

func main() {

	if false {
		contexts := make([]context.Context, 0, 3)
		for i := 0; i < 3; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			contexts = append(contexts, ctx)
		}
		for i, ctx := range contexts {
			//fmt.Printf("i=%d ctx=%+v \n", i, ctx)
			fmt.Printf("i=%d ctx=%+v Err:%+v \n", i, ctx, ctx.Err())
			//if ctx.Err() != nil {
			//fmt.Println("Go routine ", i, "canceled due to", (*ctx).Err())
			//}
		}
	}

	// waitgroup和context结合，模拟并发超时
	// 如果其中1个超时，则废弃执行,只需要返回那两个的结果
	//
	if true {
		way := 4
		if way == 1 {
			timeout := 3 * time.Second
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			var wg sync.WaitGroup
			wg.Add(3)
			var res1, res2, res3 int
			go func(pg *sync.WaitGroup, res *int, ctx context.Context) {
				resCh := make(chan int)
				go op1(ctx, resCh)
				select {
				case <-ctx.Done():
					pg.Done()
				case *res = <-resCh:
					pg.Done()
				}
			}(&wg, &res1, ctx)
			go func(pg *sync.WaitGroup, res *int, ctx context.Context) {
				resCh := make(chan int)
				go op2(ctx, resCh)
				select {
				case <-ctx.Done():
					pg.Done()
				case *res = <-resCh:
					pg.Done()
				}
			}(&wg, &res2, ctx)
			go func(pg *sync.WaitGroup, res *int, ctx context.Context) {
				defer pg.Done()
				resCh := make(chan int)
				go op3(ctx, resCh)
				select {
				case <-ctx.Done():
					pg.Done()
				case *res = <-resCh:
					pg.Done()
				}
			}(&wg, &res3, ctx)
			wg.Wait()
			fmt.Println("res1=", res1, " res2=", res2, " res3=", res3)
		} else if way == 2 {
			// https://stackoverflow.com/questions/32840687/timeout-for-waitgroup-wait
			// 缺点： 真的超时了，不清楚是哪一个超时，只能通过res1,res2,res3的值，来判断谁超时了
			timeout := 3 * time.Second
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			var wg sync.WaitGroup
			wg.Add(3)
			var res1, res2, res3 int
			go func(pg *sync.WaitGroup, res *int) {
				defer pg.Done()
				*res = opp1(ctx)
			}(&wg, &res1)
			go func(pg *sync.WaitGroup, res *int) {
				defer pg.Done()
				*res = opp2(ctx)
			}(&wg, &res2)
			go func(pg *sync.WaitGroup, res *int) {
				defer pg.Done()
				*res = opp3(ctx)
			}(&wg, &res3)
			c := make(chan struct{})
			go func() {
				wg.Wait()
				close(c)
			}()
			select {
			case <-ctx.Done():
				fmt.Printf("timeout: %+v\n", ctx.Err())
				//   timed out
			case <-c:
				fmt.Printf("completed normally")
			}
			fmt.Println("res1=", res1, " res2=", res2, " res3=", res3)
		} else if way == 3 {
			// https://github.com/shomali11/parallelizer/blob/master/group.go
			// way2的封装版，加上协程池

		} else if way == 4 {
			// 核心： 不要公用一个context，把context拆开到每个go协程里面，在每个go协程里面用select去处理超时
			// 和way1很像，区别在于way1公用1个父亲context，而这里每个G程是单独生成新的子context。
			// 其实从context原理来说都一样，因为Background引用的都是全局的context，所以并无差别
			// way4把way1简单封装了一下
			// 缺点： way1和way4的方法比way2的G程数目多了一倍。
			// https://stackoverflow.com/questions/46509979/golang-waitgroup-timeout-error-handling
			elements := make([]*Element, 0, 3) //相当于res1 res2 res3
			elements = append(elements, &Element{JobType: 1, JobRes: 0})
			elements = append(elements, &Element{JobType: 2, JobRes: 0})
			elements = append(elements, &Element{JobType: 3, JobRes: 0})
			fmt.Println("len=", len(elements), " cap=", cap(elements))
			//contexts := make([]context.Context, 0, len(elements))
			var wg sync.WaitGroup
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			for _, element := range elements {
				wg.Add(1)
				go doWork(ctx, element, &wg)
			}
			wg.Wait()
			for _, e := range elements {
				fmt.Printf("Res: %+v\n", *e)
			}
		}
	}

	// WaitGroup 的坑
	if false {
		var n sync.WaitGroup
		ic := make(chan int)
		n.Add(1) //这个add非常重要，这里不add的话，可能下面的go线程还没来得及add(1)，然后就到了wait()直接退出了
		go func(n *sync.WaitGroup) {
			defer n.Done()
			for i := 0; i < 3; i++ {
				n.Add(1)
				go func(i int, n *sync.WaitGroup, ic chan<- int) {
					defer n.Done()
					ic <- i
				}(i, n, ic)
			}
		}(&n)

		go func() {
			n.Wait()
			fmt.Println("close ch")
			close(ic)
		}()
		for i := range ic {
			fmt.Println(i)
		}
	}

	if false {
		//sem用于控制并行的数量，如果是1的话，则是顺序打印数字
		//如果是2的话，则每次并发2个go去打印，只有这2个go结束后，才开始下一波
		//因此不存在协程数目暴涨的情况，sem可用于控制协程数目
		sem := make(chan struct{}, 2)
		wg := new(sync.WaitGroup)

		for i := 0; i < 100; i++ {
			sem <- struct{}{}
			wg.Add(1)
			go func(id int) {
				defer func() {
					<-sem
					wg.Done()
				}()

				process(id)
			}(i)
		}
		wg.Wait()
	}

	if false {
		wg := new(sync.WaitGroup)
		fns := make(chan func(), 2)
		//fns目的是为了控制write_db函数的并发粒度，因为如果批量写DB，则会对DB造成压力。2去掉变成无缓冲的chan也都没问题
		sem := make(chan struct{}, 3)
		//sem是为了控制for循环的粒度，

		go func() {
			for i := 0; i < 30; i++ {
				sem <- struct{}{}
				wg.Add(1)

				go func(id int) {
					defer func() {
						<-sem
					}()

					process(id)
					fns <- func() {
						write_db(id)
					}
				}(i)

			}
			wg.Wait()
			close(fns)
		}()

		for fn := range fns {
			fn()
			wg.Done()
		}
	}
}

/*
近日发现的颇有用的做法，特记录分享之。
场景很简单，读数据库，做一些处理，再写回数据库，伪码如下
//https://www.golang123.com/topic/1596
func main() {
	sem := make(chan struct{}, 32)
	wg := new(sync.WaitGroup)

	for read_db_rows() {

		sem <- struct{}{}
		wg.Add(1)

		go func() {

			defer func() {
				<-sem
				wg.Done()
			}()

			process_row()
			begin_db_transaction()
			write_db()
			commit_db_transaction()

		}()

	}
	wg.Wait()
}

就是基本的并行处理。但是每行都起一个数据库事务，显然不太经济，可以合并一下。过去我会用一个chan Result来做，现在发现用chan func()可以更直观：

func main() {
	wg := new(sync.WaitGroup)
	fns := make(chan func(), 2048)
	sem := make(chan struct{}, 32)

	go func() {
		for read_db_rows() {

			sem <- struct{}{}
			wg.Add(1)

			go func() {

				defer func() {
					<-sem
				}()

				process_row()
				fns <- func() {
					write_db()
				}

			}()

		}
		wg.Wait()
		close(fns)
	}()

	n := 0
	begin_db_transaction()

	for fn := range fns {
		fn()
		n++
		if n == 512 {
			commit_db_transaction()
			begin_db_transaction()
			n = 0
		}
		wg.Done()
	}

	commit_db_transaction()
}

和go官方博客里描述的pipeline不同（https://blog.golang.org/pipelines），这里的fns是有缓冲的。
没有缓冲也可以达到合并事务的效果，但是process_row和write_db的执行时间不一样，加个缓冲可以在其中一个过程发生抖动时，不影响另一个的吞吐。

事务的开始和提交，以及wait group的Done方法，都放到了另一个goroutine里，read_db_rows()循环可以全速跑而不受write_db()的影响。

整个pipeline的终止方式和go博客里的一样，直接close用于通讯的chan。因为close发生于wg.Wait之后，下一级已经确定处理完所有行了，所以就算是带缓冲的chan也可以直接close。

chan的缓冲大小可以根据实际调整。我这里在数据库抖动时，可能积累到几千个func()，所以实际用的是10万，不怕大就怕小，
如果太小，fns <- func() { ... }就阻塞了，影响循环【因为有defer <-sem,只有把func加入到fns才能执行defer，释放sem，让for继续】。事务合并的计数也不一定是512，要看总的吞吐。

这种做法的好处是，pipeline上游传给下游的变量，都被func()捕捉了，不需要定义一个类型并用chan来传递，简洁很多。
*/
