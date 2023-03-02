package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

func main() {
	// 创建连接池
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	// 创建客户端
	client := &http.Client{
		Timeout:   time.Second * 30,
		Transport: transport,
	}
	// 发起请求  Get请求
	resp, err := client.Get("http://127.0.0.1:8080/hello")
	if err != nil {
		panic(err)
	}
	// 关闭连接
	defer resp.Body.Close()

	// 处理服务器响应
	bds, _ := io.ReadAll(resp.Body)
	fmt.Println(string(bds))

}
