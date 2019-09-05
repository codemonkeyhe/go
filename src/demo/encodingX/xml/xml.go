//https://www.cnblogs.com/jkko123/p/8325813.html

package main

import (
	"encoding/xml"
	"fmt"
)

//我们通过定义一个结构体，来解析xml
//注意，结构体中的字段必须是可导出的
type Books struct {
	//如果有类型为xml.Name的XMLName字段，则解析时会保存元素名到该字段
	XMLName xml.Name `xml:"books"`
	//定义的字段中包含,attr，则解析时会把对应字段的属性值赋给该字段
	Nums int `xml:"nums,attr"`
	//定义的字段名含有xml中元素名，则解析时会把该元素值赋给该字段
	//Bo []Book `xml:"book"`
	Bo Book `xml:"book"` //也就是子节点不一定采用 > >的模式，可以用xml.Name来代替

	//字段类型为string或[]byte，并且包含,innerxml，则解析时会把此字段对应的元素内所有原始xml累加到字段上
	Data string `xml:",innerxml"`

	//字段定义包含-，则解析时不会为该字段匹配任何数据
	Tmp int `xml:"-"`
}

type Book struct {
	XMLName xml.Name `xml:"book"`
	Name    string   `xml:"name,attr"`
	Author  string   `xml:"author"`
	Time    string   `xml:"time"`

	//字段定义如a>b>c，这样，解析时会从xml当前节点向下寻找元素并将值赋给该字段
	Types []string `xml:"types>type"`

	//字段定义有,any，则解析时如果xml元素没有与任何字段匹配，那么这个元素就会映射到该字段
	Test string `xml:",any"`
}

func main() {
	//xml数据字符串
	data := `<?xml version="1.0" encoding="utf-8"?>
            <books nums="2">
                <book name="思想">
                    <author>小张</author>
                    <time>2018年1月20日</time>
                    <types>
                        <type>教育</type>
                        <type>历史</type>
                    </types>
                    <test>我是多余的</test>
                </book>
                <book name="政治">
                    <author>小王</author>
                    <time>2018年1月20日</time>
                    <types>
                        <type>科学</type>
                        <type>人文</type>
                    </types>
                    <test>我是多余的</test>
                </book>
            </books>`

	//创建一个Books对象
	bs := Books{}
	//把xml数据解析成bs对象
	xml.Unmarshal([]byte(data), &bs)
	//打印bs对象中数据
	fmt.Println(bs.XMLName)
	fmt.Println(bs.Nums)
	fmt.Println(bs.Tmp)
	fmt.Printf("===%+v", bs)
	/*
			 //循环打印Book
		    for _, v := range bs.Book {
		        fmt.Println(v.XMLName);
		        fmt.Println(v.Name);
		        fmt.Println(v.Author);
		        fmt.Println(v.Time);
		        fmt.Println(v.Types);
		        fmt.Println(v.Test);
		    }*/
}
