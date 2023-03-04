package main

import (
	"fmt"
	"net"
	"net/rpc"
)

// 服务端注册rpc服务，给对象绑定方法
type Hello struct {
}

func (h *Hello) SayHello(req string, rep *string) error {
	*rep = req + " hello!"
	return nil
}

func main() {
	// 注册rpc服务，绑定对象方法
	// 服务名称： SayHello 处理器： Hello
	err := rpc.RegisterName("SayHello", &Hello{})
	if err != nil {
		fmt.Println("注册 rpc 服务失败!", err)
		return
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
	rpc.ServeConn(conn)
}
