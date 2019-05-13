package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

/*

Golang中如何让html/template不转义html标签
http://blog.xiayf.cn/2013/11/01/unescape-html-in-golang-html_template/

https://stackoverflow.com/questions/36931742/golang-dont-escape-in-html-templates
https://stackoverflow.com/questions/14765395/why-am-i-seeing-zgotmplz-in-my-go-html-template-output/33201432


*/

//https://colobu.com/2016/10/09/Go-embedded-template-best-practices/
func handler1(w http.ResponseWriter, r *http.Request) {
	//必须是模板使用在前template，模板定义在后define，因此下面两个文件顺序不能交换
	t, _ := template.ParseFiles("header.html", "footer.html")
	err := t.Execute(w, map[string]string{"Title": "My title", "Body": "Hi this is my body"})
	if err != nil {
		panic(err)
	}
}

//https://golang.org/src/text/template/example_test.go
func handler2(w http.ResponseWriter, r *http.Request) {
	const letter = `
Dear {{.Name}},
{{if .Attended}}
It was a pleasure to see you at the wedding.
{{- else}}
It is a shame you couldn't make it to the wedding.
{{- end}}
{{with .Gift -}}
Thank you for the lovely {{.}}.
{{end}}
Best wishes,
Josie
`
	// Prepare some data to insert into the template.
	type Recipient struct {
		Name, Gift string
		Attended   bool
	}
	var recipients = []Recipient{
		{"Aunt Mildred", "bone china tea set", true},
		{"Uncle John", "moleskin pants", false},
		{"Cousin Rodney", "", false},
	}

	// Create a new template and parse the letter into it.
	t := template.Must(template.New("letter").Parse(letter))

	// Execute the template for each recipient.
	for _, data := range recipients {
		//err := t.Execute(os.Stdout, data)
		err := t.Execute(w, data)
		if err != nil {
			log.Println("executing template:", err)
		}
	}

	fmt.Println("handler2 done")
}

type Friend struct {
	Fname string
}

type Person struct {
	UserName string
	Emails   []string
	Friends  []*Friend
}

func EmailDealWith(args ...interface{}) string {
	ok := false
	var s string
	if len(args) == 1 {
		s, ok = args[0].(string)
	}
	if !ok {
		s = fmt.Sprint(args...)
	}
	// find the @ symbol
	substrs := strings.Split(s, "@")
	if len(substrs) != 2 {
		return s
	}
	// replace the @ by " at "
	return (substrs[0] + " at " + substrs[1])
}

func ShowTime(t time.Time, format string) string {
	return t.Format(format)
}

type User struct {
	Username, Password string
	RegTime            time.Time
}

/*
{{"string"}} // 一般 string
{{`raw string`}} // 原始 string
{{'c'}} // byte
{{print nil}} // nil 也被支持

type HTML
type HTMLAttr
type JS
type JSStr
type URL
type CSS

*/

func copyrightYear() string {
	return fmt.Sprintf("%d", time.Now().Year())
}

type Bulletin struct {
	Appid   uint32
	Img     string
	Content template.JSStr
	Name    string
	Ico     string
}

type IndexData struct {
	BtData []*Bulletin
}

func main() {
	fmt.Println("hhhhhhhhhhh")

	if false {
		var err error
		data := template.JS(`var appObject30032 = new Object();
        appObject30032.img = 'http://dl.open.yy.com/yy5/images/30032_big.png';
        appObject30032.content = '<p>游戏套件</p><p>排麦召唤队友 秀出游戏实力</p>';
        appObject30032.name = '游戏套件';
        appObject30032.ico = 'http://dl.open.yy.com/yy5/images/30032_small.png';
        appArray['30032'] = appObject30032;`)
		tpl, _ := template.New("").Parse(`<script type="text/javascript">{{.}}</script>`)
		err = tpl.Execute(os.Stdout, data)
		if err != nil {
			fmt.Printf("%+v", err)
		}
	}

	if true {
		var err error
		tpl, _ := template.New("").Parse(`
		<script type="text/javascript">
var appArray=new Array();
{{with .BtData}}{{range .}}
var appObject{{.Appid}} = new Object();
appObject{{.Appid }}.img = '{{.Img}}';
appObject{{.Appid}}.content = '{{.Content}}';
appArray['{{.Appid}}'] = appObject{{.Appid}};
{{end}}{{end}}
</script>
		`)
		data := &IndexData{
			BtData: make([]*Bulletin, 0, 2),
		}
		bt := &Bulletin{
			Appid:   30032,
			Img:     "http://dl.open.yy.com/yy5/images/30032_big.png",
			Content: template.JSStr("<p>游戏套件</p><p>排麦召唤队友 秀出游戏实力</p>"),
		}
		data.BtData = append(data.BtData, bt)
		err = tpl.Execute(os.Stdout, data)
		if err != nil {
			fmt.Printf("%+v", err)
		}
	}

	if false { //属性转换1
		var err error
		tpl, _ := template.New("").Parse(`<a {{.AAA}}>XX</a>`)
		var data map[string]interface{} = make(map[string]interface{})
		data["AAA"] = template.HTMLAttr(`data-json='{"Hello":"World"}'`)
		//err = tpl.Execute(os.Stdout, data)
		err = tpl.Execute(os.Stdout, nil)
		if err != nil {
			fmt.Printf("%+v", err)
		}
	}

	if false { //属性转换2 完全行不通
		var err error
		tpl, _ := template.New("").Parse(`<a data-json={{.AAA}}>XX</a>`)
		var data map[string]interface{} = make(map[string]interface{})
		//data["AAA"] = template.URL(`'{"Hello":"World"}'`)
		//Out:  <a data-json=&#39;{&#34;Hello&#34;:&#34;World&#34;}&#39;>XX</a>
		//data["AAA"] = template.JSStr(`'{"Hello":"World"}'`)
		//Out:  <a data-json=&#39;{&#34;Hello&#34;:&#34;World&#34;}&#39;>XX</a>
		// data["AAA"] = template.JS(`'{"Hello":"World"}'`)
		//Out:  <a data-json=&#39;{&#34;Hello&#34;:&#34;World&#34;}&#39;>XX</a>

		//data["AAA"] = template.HTML(`'{"Hello":"World"}'`)
		//Out:  <a data-json=&#39;{&#34;Hello&#34;:&#34;World&#34;}&#39;>XX</a>
		data["AAA"] = template.HTMLAttr(`'{"Hello":"World"}'`)
		//Out:  <a data-json=&#39;{&#34;Hello&#34;:&#34;World&#34;}&#39;>XX</a>
		err = tpl.Execute(os.Stdout, data)
		if err != nil {
			fmt.Printf("%+v", err)
		}
	}

	if false { //属性转换3 完全行不通
		var err error
		tpl, _ := template.New("").Parse(`<a data-json{{.AAA}}>XX</a>`)
		var data map[string]interface{} = make(map[string]interface{})
		//data["AAA"] = template.URL(`='{"Hello":"World"}'`)
		//Out:  <a data-jsonZgotmplZ>XX</a>
		//data["AAA"] = template.JSStr(`='{"Hello":"World"}'`)
		//Out: <a data-jsonZgotmplZ>XX</a>
		//data["AAA"] = template.JS(`='{"Hello":"World"}'`)
		//Out:  <a data-jsonZgotmplZ>XX</a>

		//data["AAA"] = template.HTML(`='{"Hello":"World"}'`)
		//Out: <a data-jsonZgotmplZ>XX</a>
		data["AAA"] = template.HTMLAttr(`='{"Hello":"World"}'`)
		//Out:  <a data-json='{"Hello":"World"}'>XX</a>
		err = tpl.Execute(os.Stdout, data)
		if err != nil {
			fmt.Printf("%+v", err)
		}
	}

	if false { //使用自定义函数
		fm := template.FuncMap{
			"copyrightYear": func() string {
				return fmt.Sprintf("%d", time.Now().Year())
			},
		}
		tmp := template.Must(template.New("").Funcs(fm).Parse("{{copyrightYear}}"))
		tmp.Execute(os.Stdout, nil)
	}

	if false { //模板名必须与文件同名
		// https://stackoverflow.com/questions/49043292/error-template-is-an-incomplete-or-empty-template/49043639
		//		fm := template.FuncMap{
		//			"copyrightYear": func() string {
		//				return fmt.Sprintf("%d", time.Now().Year())
		//			},
		//		}
		fm := template.FuncMap{
			"copyrightYear": copyrightYear,
		}
		tmp := template.New("index.html").Funcs(fm)
		tmp, err := tmp.ParseFiles("./index.html")
		//错误做法
		//tmp, err := template.ParseFiles("./index.html")

		//正确做法
		//tmp, err := template.New("index.html").Funcs(fm).ParseFiles("./index.html")
		//tmp, err := template.New("index.html").Funcs(fm).ParseFiles("index.html")
		//下面是错误的做法
		//tmp, err := template.New("index").Funcs(fm).ParseFiles("index.html")
		//tmp, err := template.New("").Funcs(fm).ParseFiles("index.html")

		if err != nil {
			fmt.Println(err)
		}
		if tmp == nil {
			fmt.Printf("------init load Failed!Tpl: %+v\n", tmp)
		}
		fmt.Printf("%+v\n", tmp)
		err = tmp.Execute(os.Stdout, nil)
		if err != nil {
			fmt.Println(err)
		}
	}

	if false { //error case【实际不会报错】
		t := template.New("fieldname example")
		t, _ = t.Parse("hello {{.UserName}}! {{.email}}")
		p := Person{UserName: "Astaxie"}
		t.Execute(os.Stdout, p)
		//上面的代码就会报错，因为我们调用了一个未导出的字段，但是如果我们调用了一个不存在的字段是不会报错的，而是输出为空。
	}

	if false {
		//https://github.com/astaxie/build-web-application-with-golang/blob/master/zh/07.4.md
		//Go语言的模板通过{{}}来包含需要在渲染时被替换的字段，{{.}}表示当前的对象，这和Java或者C++中的this类似，
		//如果要访问当前对象的字段通过{{.FieldName}}，但是需要注意一点：这个字段必须是导出的(字段首字母必须是大写的)，否则在渲染的时候就会报错
		t := template.New("fieldname example")
		t, _ = t.Parse("hello {{.UserName}}!")
		p := Person{UserName: "Astaxie"}
		t.Execute(os.Stdout, p)
	}

	if false {
		/*
			可以使用{{with …}}…{{end}}和{{range …}}{{end}}来进行数据的输出。

			{{range}} 这个和Go语法里面的range类似，循环操作数据
			{{with}}操作是指当前对象的值，类似上下文的概念
		*/

		f1 := Friend{Fname: "minux.ma"}
		f2 := Friend{Fname: "xushiwei"}
		t := template.New("fieldname example")
		t, _ = t.Parse(`hello {{.UserName}}!
			{{range .Emails}}
				an email {{.}}
			{{end}}
			{{with .Friends}}
			{{range .}}
				my friend name is {{.Fname}}
			{{end}}
			{{end}}
			`)
		p := Person{UserName: "Astaxie",
			Emails:  []string{"<astaxie@beego.me", "astaxie@gmail.com"},
			Friends: []*Friend{&f1, &f2}}
		t.Execute(os.Stdout, p)
	}

	if false {
		/*
			每一个模板函数都有一个唯一值的名字，然后与一个Go函数关联，通过如下的方式来关联
			type FuncMap map[string]interface{}
			例如，如果我们想要的email函数的模板函数名是emailDeal，它关联的Go函数名称是EmailDealWith,那么我们可以通过下面的方式来注册这个函数
			t = t.Funcs(template.FuncMap{"emailDeal": EmailDealWith})
			EmailDealWith这个函数的参数和返回值定义如下：
			func EmailDealWith(args …interface{}) string

			an emails {{.|emailDeal}}改为
			an emails {{.|html}}
		*/
		f1 := Friend{Fname: "minux.ma"}
		f2 := Friend{Fname: "xushiwei"}
		t := template.New("fieldname example")
		t = t.Funcs(template.FuncMap{"emailDeal": EmailDealWith})
		t, _ = t.Parse(`hello {{.UserName}}!
				{{range .Emails}}
					an emails {{.|html}}
				{{end}}
				{{with .Friends}}
				{{range .}}
					my friend name is {{.Fname}}
				{{end}}
				{{end}}
				`)
		p := Person{UserName: "Astaxie",
			Emails:  []string{"astaxie@beego.me", "astaxie@gmail.com"},
			Friends: []*Friend{&f1, &f2}}
		t.Execute(os.Stdout, p)
		/*模板包内部已经有内置的模板函数

		var builtins = FuncMap{
			"and":      and,
			"call":     call,
			"html":     HTMLEscaper,
			"index":    index,
			"js":       JSEscaper,
			"len":      length,
			"not":      not,
			"or":       or,
			"print":    fmt.Sprint,
			"printf":   fmt.Sprintf,
			"println":  fmt.Sprintln,
			"urlquery": URLQueryEscaper,
		}
		{{. | html}}  在email输出的地方我们可以采用如上方式可以把输出全部转化html的实体

		*/
	}

	if false {
		//https://www.teakki.com/p/57df64d5da84a0c4533815be
		// https://cloud.tencent.com/developer/article/1072981
		//showtime是自定义的模板函数 {{FuncName arg1 arg2}}
		//.Format是系统内置函数
		u := User{"dotcoo", "dotcoopwd", time.Now()}
		t, err := template.New("text").Funcs(template.FuncMap{"showtime": ShowTime}).
			Parse(`<p>{{.Username}}|{{.Password}}|{{.RegTime.Format "2006-01-02 15:04:05"}}</p>
<p>{{.Username}}|{{.Password}}|{{showtime .RegTime "2006-01-02 15:04:05"}}</p>
`)
		if err != nil {
			panic(err)
		}
		t.Execute(os.Stdout, u)
	}

	http.HandleFunc("/1", handler1)
	http.HandleFunc("/2", handler2)
	http.ListenAndServe(":8080", nil)

	//下面这句不会执行，除非编译成exe，在window下点击app执行
	fmt.Println("no chance to show as ListenAndServe before!")

	//用于windows执行时显示stdout信息
	fmt.Scanln()
}
