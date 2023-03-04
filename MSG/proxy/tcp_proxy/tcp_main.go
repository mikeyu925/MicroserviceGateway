package main

import (
	"MSG/proxy/tcp_proxy/proxy"
	"MSG/proxy/tcp_proxy/server"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// 定义自己想要实现的handler，因为Handler是接口类型
type myHandler struct {
}

// 实现 TCPHandler 对应的方法
func (h *myHandler) ServerTCP(conn net.Conn, ctx context.Context) {
	// 接收数据
	buffer := make([]byte, 1024)
	n, _ := conn.Read(buffer)
	fmt.Println("收到了一个请求:", string(buffer[:n]))
	conn.Write([]byte("这里是我自己定义的myhandler 我们收到了你的TCP请求！\n"))
}

func main() {
	// 启动tcp服务器协程
	go func() {
		var addr = "127.0.0.1:8003"
		// 创建TCPServer实例
		tcpServer := &server.TCPServer{
			Addr:    addr,
			Handler: &myHandler{},
		}
		fmt.Println("启动TCP Server...")
		// 启动监听提供服务
		tcpServer.ListenAndServe()
	}()

	time.Sleep(time.Second * 1)
	// 启动tcp代理服务器协程
	go func() {
		// 下游服务器地址
		var tcpServerAddr = "127.0.0.1:8003"
		// 创建TCPProxy实例
		tcpProxy := proxy.NewSingleHostReverseProxy(tcpServerAddr)
		fmt.Println("启动TCP ProxyServer...")
		var proxyAddr = "127.0.0.1:8083"
		// 启动监听提供服务
		log.Fatal(server.ListenAndServe(proxyAddr, tcpProxy))
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // 监控系统接收和终止命令
	<-quit
}
