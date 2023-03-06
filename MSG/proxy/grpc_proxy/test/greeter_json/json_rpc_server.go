package main

import (
	"MSG/proxy/grpc_proxy/test/greeter_json/inters"
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
)

// 服务端注册rpc服务，给对象绑定方法
type Hello struct {
}

func (h *Hello) SayHello(req string, rep *string) error {
	*rep = req + " 你好!"
	return nil
}

func RegisterService(handler inters.Myinterface) error {
	err := rpc.RegisterName(inters.HelloServiceName, handler)
	if err != nil {
		return errors.New("注册 rpc 服务失败!")
	}
	return nil
}

func main() {
	// 注册rpc服务，绑定对象方法
	// 服务名称： SayHello 处理器： Hello
	err := RegisterService(&Hello{})
	if err != nil {
		fmt.Println("注册 rpc 服务失败!")
	}
	// 创建设置监听
	listener, err := net.Listen("tcp", "127.0.0.1:8004")
	if err != nil {
		fmt.Println("监听失败!", err)
		return
	}
	fmt.Println("listening port: ", "127.0.0.1:8004")
	// 建立连接
	conn, err := listener.Accept()
	if err != nil {
		fmt.Println("创建连接失败!", err)
		return
	}
	fmt.Println("connection accept...")
	// 绑定服务：将连接绑定rpc服务
	jsonrpc.ServeConn(conn)
}
