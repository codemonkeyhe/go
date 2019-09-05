package main

import (
	"context"
	"fmt"

	proto "github.com/micro/examples/service/proto"
	micro "github.com/micro/go-micro"
)

/*

https://micro.mu/docs/go-micro.html
https://github.com/micro/examples

https://github.com/micro/examples/tree/master/service
https://github.com/micro/examples/tree/master/function

https://github.com/micro/examples/tree/master/pubsub



*/

//服务端的实现逻辑
type Greeter struct{}

func (g *Greeter) Hello(ctx context.Context, req *proto.HelloRequest, rsp *proto.HelloResponse) error {
	rsp.Greeting = "Hello " + req.Name
	return nil
}

//Create a message handler. It’s signature should be func(context.Context, v interface{}) error.
func ProcessEvent(ctx context.Context, event *proto.Event) error {
	fmt.Printf("Got event %+v\n", event)
	return nil
}

func main() {
	runServer := true
	runClient := true
	runFunction := false
	runPublish := false
	runSubscribe := false
	if runServer {
		// Create a new service. Optionally include some options here.
		service := micro.NewService(
			micro.Name("greeter"),
		)

		// Init will parse the command line flags.
		service.Init()

		// Register handler
		proto.RegisterGreeterHandler(service.Server(), new(Greeter))

		// Run the server
		if err := service.Run(); err != nil {
			fmt.Println(err)
		}
	}

	if runClient {
		// Create a new service. Optionally include some options here.
		service := micro.NewService(micro.Name("greeter.client"))
		service.Init()

		// Create new greeter client
		greeter := proto.NewGreeterService("greeter", service.Client())

		// Call the greeter
		rsp, err := greeter.Hello(context.TODO(), &proto.HelloRequest{Name: "John"})
		if err != nil {
			fmt.Println(err)
		}

		// Print response
		fmt.Println(rsp.Greeting)
	}
	if runFunction {
		/*
			Go Micro includes the Function programming model.
			A Function is a one time executing Service which exits after completing a request
		*/
		// create a new function
		fnc := micro.NewFunction(
			micro.Name("greeter"),
		)

		// init the command line
		fnc.Init()

		// register a handler
		fnc.Handle(new(Greeter))

		// run the function
		fnc.Run()
	}
	if runPublish {
		//Create a new publisher with a topic name and service client
		p := micro.NewPublisher("events", service.Client())
		//Publish a proto message
		p.Publish(context.TODO(), &proto.Event{Name: "event"})
	}
	if runSubscribe {

		//Register the message handler with a topic
		micro.RegisterSubscriber("events", ProcessEvent)
	}
}
