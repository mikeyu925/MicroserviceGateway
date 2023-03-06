package inters

// 服务名
const HelloServiceName = "Hello"

// 服务方法
const HelloServiceMethod = "Hello.SayHello"

// 相当于服务
type Myinterface interface {
	SayHello(args string, reply *string) error
}
