package main

import (
	"MSG/middleware/servicediscovery/zookeeper"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	server1 := RealServer{"127.0.0.1:8001"}
	server1.Run()
	server2 := RealServer{"127.0.0.1:8002"}
	server2.Run()

	waitChan := make(chan os.Signal)
	signal.Notify(waitChan, syscall.SIGINT, syscall.SIGTERM) // 监听系统的关闭信号 Ctrl + C 或者 Kill
	<-waitChan
}

// 下游真实服务器
type RealServer struct {
	Addr string // 服务器主机地址 ip:port
}

const (
	RegisterAddr string = "192.168.153.132:2181"
)

func (r *RealServer) Run() {
	mux := http.NewServeMux()
	mux.HandleFunc("/realserver", r.HelloHandler)
	// 创建一个server
	server := &http.Server{
		Addr:         r.Addr,
		Handler:      mux,
		WriteTimeout: time.Second * 3,
	}
	go func() {
		// 将服务注册到注册中心
		zkManager := zookeeper.NewZkManager([]string{""})
		err := zkManager.GetConnect()
		if err != nil {
			fmt.Sprintf("connect zookeeper eror : %v", err.Error())
		}
		defer zkManager.Close()
		err = zkManager.RegisterServerPath("/realserver", fmt.Sprintf(r.Addr))
		if err != nil {
			fmt.Println("register node error : ", err.Error())
		}
		log.Fatal(server.ListenAndServe())
	}()
}

// 路由处理器
func (r *RealServer) HelloHandler(w http.ResponseWriter, req *http.Request) {
	newPath := fmt.Sprintf("Here is real server : http://%s %s", r.Addr, req.URL.Path)
	w.Write([]byte(newPath))
}
