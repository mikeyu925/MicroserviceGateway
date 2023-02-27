package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {
	fmt.Println("正向代理服务器启动, Port: 8080")
	http.Handle("/", &Pxy{})
	http.ListenAndServe("127.0.0.1:8080", nil)
}

type Pxy struct {
}

/*
// 实现了 Handler 的方法

	type Handler interface {
		ServeHTTP(ResponseWriter, *Request)
	}
*/
func (p *Pxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// 1. 代理服务器接收客户端请求，并封装为新的请求
	outReq := &http.Request{}
	*outReq = *req
	// 2. 发送请求到下游真实服务器，接收响应
	transport := http.DefaultTransport // 创建连接池
	resp, err := transport.RoundTrip(outReq)
	if err != nil {
		rw.WriteHeader(http.StatusBadGateway) // 返回响应给客户端
		return
	}
	// 3. 处理响应并返回上游客户端
	// 把下游服务器所有头信息进行拷贝
	for key, value := range resp.Header {
		for _, v := range value {
			rw.Header().Add(key, v)
		}
	}
	rw.WriteHeader(resp.StatusCode)
	io.Copy(rw, resp.Body)
	resp.Body.Close()
}
