package main

import "net/http"

func main() {
	// 注册路由 和 回调函数
	http.HandleFunc("/hello", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("hello, this is server!"))
	})
	// 启动监听并提供服务
	http.ListenAndServe("127.0.0.1:8080", nil)
}
