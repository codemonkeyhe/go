package main

import (
	"testing"
)

var a1 []int
var a2 []int
var a3 []int

/*
[monkey@bogon merge]$go test -bench=. -benchmem
goos: linux
goarch: amd64
pkg: demo/merge
BenchmarkMergeBetter-4   	  500000	      2924 ns/op	     224 B/op	       4 allocs/op
BenchmarkMergeNormal-4   	30000000	        40.6 ns/op	      80 B/op	       1 allocs/op
PASS
ok  	demo/merge	2.761s
[monkey@bogon merge]$

[monkey@bogon merge]$go test -bench=. -benchmem
goos: linux
goarch: amd64
pkg: demo/merge
BenchmarkMergeBetter-4   	  200000	      7819 ns/op	     464 B/op	       4 allocs/op
BenchmarkMergeNormal-4   	20000000	        93.2 ns/op	     320 B/op	       1 allocs/op
PASS
ok  	demo/merge	3.611s


结论： BenchmarkMergeBetter表现不及预期，因为内存分配过多，还不如内存分配少的快
另外的原因是，BenchmarkMergeNormal虽然是串行的合并，但是数据量非常小，相当于并行的优势发挥不出来。可能要到成千上万的slice才能体现出并行的优势。
因此，可以见得，在查找搜索这些过程，当数据量小的时候，直接遍历反而更快。否则，要构造去重的集合map[int]struct{}多了内存分配和数据插入的操作，反而会更加慢
*/

func init() {
	//a1 = []int{1, 3, 5, 7}
	//a2 = []int{2, 4, 6, 8, 0, 1}
	a1 = []int{1, 3, 5, 7, 2, 4, 6, 8, 0, 2, 4, 6, 8, 0, 2, 4, 6, 8, 0}
	a2 = []int{2, 4, 6, 8, 0, 1, 11, 45, 33, 78, 11, 45, 33, 78, 11, 45, 33, 78}
	a3 = []int{11, 45, 33, 78}
}

func BenchmarkMergeNormalCopy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		arrayMergeNormalCopy(a1, a2)
	}
}

func BenchmarkMergeNormal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		arrayMergeNormal(a1, a2)
	}
}

func BenchmarkMergeBetter(b *testing.B) {
	for i := 0; i < b.N; i++ {
		arrayMergeBetter(a1, a2)
	}
}

func BenchmarkMergeBetterV2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		arrayMergeBetterV2(a1, a2)
	}
}
