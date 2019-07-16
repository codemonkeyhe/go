package main

import (
	"fmt"
	"sync"
)

func process(id int) {
	fmt.Println(id)
}

func write_db(id int) {
	fmt.Printf("Write:%d\n", id)
}

func main() {

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

	if true {
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
