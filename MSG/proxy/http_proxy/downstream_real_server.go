package http_proxy

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	server1 := RealServer{"127.0.0.1:8001"}
	server1.Run()

	waitChan := make(chan os.Signal)
	signal.Notify(waitChan, syscall.SIGINT, syscall.SIGTERM) // 监听系统的关闭信号 Ctrl + C 或者 Kill
	<-waitChan
}

// 下游真实服务器
type RealServer struct {
	Addr string // 服务器主机地址 ip:port
}

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
		server.ListenAndServe()
	}()
}

// 路由处理器
func (r *RealServer) HelloHandler(w http.ResponseWriter, req *http.Request) {
	newPath := fmt.Sprintf("Here is real server : http://%s %s", r.Addr, req.URL.Path)
	w.Write([]byte(newPath))
}
