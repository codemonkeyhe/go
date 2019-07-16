package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

/*
海量小文件的创建 和遍历 性能测试
*/

func genFileName(dataIn string) <-chan string {
	out := make(chan string)
	go func() {
		file, err := os.Open(dataIn)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		br := bufio.NewReader(file)
		for {
			a, _, c := br.ReadLine()
			if c == io.EOF {
				break
			}
			// fmt.Println(string(a))
			out <- string(a)
		}
		close(out)
	}()
	return out
}

// 过滤文件，比如存在的不创建
func checkFile(files <-chan string) <-chan string {
	out := make(chan string)
	go func() {
		for file := range files {
			dirName := filepath.Dir(file)
			fileName := filepath.Base(file)
			fmt.Printf("file:%s dirName:%s fileName:%s\n", file, dirName, fileName)
		}
		close(out)
	}()
	return out
}

func createFile(file string) error {
	// dirName := filepath.Dir(file)
	fileName := filepath.Base(file)
	// fmt.Printf("file:%s dirName:%s fileName:%s\n", file, dirName, fileName)
	if fileName == "" || fileName == "." {
		return nil
	}
	newFile, err := os.Create(file)
	if err != nil {
		fmt.Println(err)
		return err
	}
	newFile.Close()
	return nil
}

func makeFile(files <-chan string) {
	sem := make(chan struct{}, 1000)
	for file := range files {
		sem <- struct{}{}
		go func(file string) {
			defer func() {
				<-sem
			}()
			createFile(file)
		}(file)
	}
}

//调用os.MkdirAll递归创建文件夹
func createDirAll(filePath string) error {
	if !isExist(filePath) {
		err := os.MkdirAll(filePath, os.ModePerm)
		return err
	}
	return nil
}

// 判断所给路径文件/文件夹是否存在(返回true是存在)
func isExist(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// https://blog.csdn.net/skh2015java/article/details/81531126

func createDir(dirs []string) {
	for _, dir := range dirs {
		createDirAll(dir)
	}
}

var semaWalk = make(chan struct{}, 1000)

func walkDir(dir string, n *sync.WaitGroup, file chan<- string) error {
	defer n.Done()
	semaWalk <- struct{}{}
	defer func() {
		<-semaWalk
	}()
	f, err := os.Open(dir)
	if err != nil {
		return err
	}
	list, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return err
	}
	for _, finfo := range list {
		if finfo.IsDir() {
			subdir := filepath.Join(dir, finfo.Name())
			n.Add(1)
			go walkDir(subdir, n, file)
		} else {
			file <- filepath.Join(dir, finfo.Name())
		}
	}
	return nil
}

const (
	//	SrcDir = "/home/PPPoker/monkey/tinyfile"
	SrcDir = "./"
)

func main() {

	// 1 并发遍历src_dir目录,并发粒度由sema控制
	filech := make(chan string)
	var n sync.WaitGroup
	n.Add(1)
	go walkDir(SrcDir, &n, filech)
	go func() {
		n.Wait()
		close(filech)
	}()

	for file := range filech {
		fmt.Println(file)
	}

	if false {
		dirs := []string{"origin", "head/google", "head/facebook", "robot", "guest/.svn/prop-base", "share/.svn"}
		createDir(dirs)

		dataIn := "data.in"
		fileNames := genFileName(dataIn)
		makeFile(fileNames)
	}

	//select {}
	// res := checkFile(fileNames)
	// for r := range res {
	// 	fmt.Println(r)
	// }

	// for filename := range fileNames {
	// 	fmt.Println(filename)
	// }

}
