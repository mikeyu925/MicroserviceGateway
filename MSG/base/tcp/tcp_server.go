package main

import (
	"fmt"
	"net"
)

func main() {
	// 监听服务器指定的端口
	listener, err := net.Listen("tcp", "127.0.0.1:3000")
	if err != nil {
		fmt.Println(err)
		return
	}
	// 创建TCP连接
	conn, err := listener.Accept() // 连接完成前阻塞
	if err != nil {
		fmt.Println(err)
	}
	// 释放连接
	defer conn.Close()
	// 处理客户端请求，打印数据到控制台
	buf := make([]byte, 1024)
	n, _ := conn.Read(buf) // 读取之前，阻塞
	fmt.Println("get info: ", string(buf[:n]))
	// 对客户端进行响应
	conn.Write([]byte("hello client!"))
}
