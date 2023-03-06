package main

import (
	"MSG/proxy/grpc_proxy/test/greeter_json/inters"
	"errors"
	"fmt"
	"net/rpc"
	"net/rpc/jsonrpc"
)

type MyClient struct {
	c *rpc.Client
}

func NewClient(address string) (*MyClient, error) {
	// 使用 rpc 链接服务器
	//conn, err := rpc.Dial("tcp", address) // 传输方式：gob
	conn, err := jsonrpc.Dial("tcp", address) // 传输方式：json
	if err != nil {
		return nil, errors.New("Dial err : " + err.Error())
	}
	return &MyClient{c: conn}, nil
}

// 其实就是一个实现函数，实现的就是 inter.MyInterface
func (mc *MyClient) SayHello(arg string, reply *string) error {
	// 此处的服务方法名原则是应该通过反射的方式得到的
	return mc.c.Call(inters.HelloServiceMethod, arg, &reply)
}

func main() {
	MClient, err := NewClient("127.0.0.1:8004")
	defer MClient.c.Close()
	// 调用远程函数
	var reply string
	err = MClient.SayHello("小鱼", &reply)
	if err != nil {
		fmt.Printf("Dial err : ", err)
		return
	}
	fmt.Println(reply)
}
