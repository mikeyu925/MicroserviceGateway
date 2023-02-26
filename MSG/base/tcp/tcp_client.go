package main

import (
	"fmt"
	"net"
)

func main() {
	// 与服务器建立连接
	conn, err := net.Dial("tcp", "127.0.0.1:3000")
	if err != nil {
		fmt.Println("connect failed ! err :", err)
	}
	// 关闭连接
	defer conn.Close()
	// 发送数据
	conn.Write([]byte("hello server!"))
	// 接收数据
	buffer := make([]byte, 1024)
	n, _ := conn.Read(buffer)
	fmt.Println("get info : ", string(buffer[:n]))
}
