package server

import (
	"errors"
	"fmt"
)

/*
对象的方法要能远程访问，它们必须满足一定的条件，否则这个对象的方法会被忽略
方法的类型是可输出的 (the method's type is exported)
方法本身也是可输出的 （the method is exported）
方法必须由两个参数，必须是输出类型或者是内建类型 (the method has two arguments, both exported or builtin types)
方法的第二个参数是指针类型 (the method's second argument is a pointer)
方法返回类型为 error (the method has return type error)
所以一个输出方法的格式如下：
func (t *T) MethodName(argType T1, replyType *T2) error
*/

type Args struct {
	A, B int
}

type Quotient struct {
	Quo, Rem int
}

//定义一个服务对象，这个服务对象可以很简单， 比如类型是int或者是interface{},重要的是它输出的方法。
//这里我们定义一个算术类型Arith，其实它是一个int类型，但是这个int的值我们在后面方法的实现中也没用到，所以它基本上就起一个辅助的作用。
type Arith int

func (t *Arith) Multiply(args *Args, reply *int) error {
	fmt.Printf("REQ: Args:%+v\n", args)
	*reply = args.A * args.B
	fmt.Printf("RESP: reply:%+v\n", *reply)
	return nil
}

func (t *Arith) Divide(args *Args, quo *Quotient) error {
	fmt.Printf("REQ: Args:%+v\n", args)
	if args.B == 0 {
		return errors.New("divide by zero")
	}
	quo.Quo = args.A / args.B
	quo.Rem = args.A % args.B
	fmt.Printf("RESP: Quotient:%+v\n", *quo)
	return nil
}
