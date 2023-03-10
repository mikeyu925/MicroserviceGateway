package main

import (
	"fmt"
	"net/rpc"
)

func main() {
	// 使用 rpc 链接服务器
	conn, err := rpc.Dial("tcp", "127.0.0.1:8004")
	if err != nil {
		fmt.Printf("Dial err : ", err)
		return
	}
	defer conn.Close()
	// 调用远程函数
	var reply string
	err = conn.Call("Hello.SayHello", "小鱼", &reply)
	if err != nil {
		fmt.Printf("Dial err : ", err)
		return
	}
	fmt.Println(reply)
}
