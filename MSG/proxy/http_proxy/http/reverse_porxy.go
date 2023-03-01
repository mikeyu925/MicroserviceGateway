package main

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
)

func main() {
	var port string = "8080"
	http.HandleFunc("/", handler)
	fmt.Println("反向代理服务器启动成功：" + port)
	http.ListenAndServe(":"+port, nil)
}

var (
	proxyAddr = "http://127.0.0.1:8001"
)

func handler(w http.ResponseWriter, r *http.Request) {
	// 解析下游服务器，更改请求地址 「应该是通过负载均衡策略获取，这里先采用固定代替」
	realServer, err := url.Parse(proxyAddr)
	if err != nil {
		fmt.Println("Parse url error!")
	}
	fmt.Println(realServer)
	r.URL.Scheme = realServer.Scheme // http
	r.URL.Host = realServer.Host     // ip:port
	fmt.Println(r.URL)
	// 请求下游「真实服务器」，并获取返回内容
	transport := http.DefaultTransport
	resp, err := transport.RoundTrip(r) // 得到下游服务器的响应
	if err != nil {
		fmt.Println("Request downstream server err!")
		return
	}
	defer resp.Body.Close()
	// 把下游请求内容做一些处理，然后返回给上游「客户端」
	for k, v := range resp.Header { // 修改上游响应头
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}
	bufio.NewReader(resp.Body).WriteTo(w) // 将下游响应体写回上游客户端

}
