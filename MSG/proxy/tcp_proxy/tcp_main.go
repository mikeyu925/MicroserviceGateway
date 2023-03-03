package main

import (
	"MSG/proxy/tcp_proxy/server"
	"context"
	"fmt"
	"net"
)

func main() {
	var addr = "127.0.0.1:8003"
	// 创建TCPServer实例
	tcpServer := &server.TCPServer{
		Addr:        addr,
		Handler:     &myHandler{},
		BaseContext: context.Background(),
	}
	fmt.Println("启动TCP Server...")
	// 启动监听提供服务
	tcpServer.ListenAndServe()
}

// 定义自己想要实现的handler，因为Handler是接口类型
type myHandler struct {
}

// 实现 TCPHandler 对应的方法
func (h *myHandler) ServerTCP(conn net.Conn, ctx context.Context) {

	// 接收数据
	buffer := make([]byte, 1024)
	n, _ := conn.Read(buffer)
	fmt.Println("收到了一个请求:", string(buffer[:n]))
	conn.Write([]byte("这里是 myHandler我们收到了你的TCP请求！\n"))
}
