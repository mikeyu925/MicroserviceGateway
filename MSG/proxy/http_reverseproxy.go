package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func main() {
	realSever := "http://127.0.0.1:8001?ask1=1"

	serverURL, err := url.Parse(realSever)
	if err != nil {
		fmt.Println(err)
	}
	proxy := NewSingleHostReverseProxy(serverURL)

	// 代理服务器 「采用一台主机的多个端口模仿多个主机」
	addr := "127.0.0.1:8081"
	fmt.Println("Starting proxy http server at " + addr)
	http.ListenAndServe(addr, proxy)
}

var transport = &http.Transport{
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second, // 连接超时，拨号超时时间
		KeepAlive: 30 * time.Second, // 长连接超时时间
	}).DialContext,
	MaxIdleConns:          100,              // 最大空闲连接数
	IdleConnTimeout:       90 * time.Second, // 空闲连接超时时间
	TLSHandshakeTimeout:   10 * time.Second, // tls握手超时时间
	ExpectContinueTimeout: 1 * time.Second,  // 100-continue 超时时间
}

// 多主机反向代理服务器
func NewMultipleHostsReverseProxy(ctx context.Context, targets []*url.URL) *httputil.ReverseProxy {
	// 请求协调者
	director := func(req *http.Request) {
		// 随机负载均衡：提供相同服务的URLs
		targetIndex := rand.Intn(len(targets))
		target := targets[targetIndex]
		targetQuery := target.RawQuery

		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = joinURLPath(target.Path, req.URL.Path)
		// 当对域名(非内网)反向代理时需要设置此项, 当作后端反向代理时不需要
		req.Host = target.Host
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			req.Header.Set("User-Agent", "user-agent")
		}
	}

	// 修改返回内容
	modifyFunc := func(resp *http.Response) error {
		// 兼容websocket
		if strings.Contains(resp.Header.Get("Connection"), "Upgrade") {
			// websocket协议，不需要修改返回内容
			return nil
		}

		var payload []byte
		var readErr error
		// 兼容gzip压缩
		if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
			gr, err := gzip.NewReader(resp.Body)
			if err != nil {
				return err
			}
			payload, readErr = ioutil.ReadAll(gr)
			resp.Header.Del("Content-Encoding")
		} else {
			payload, readErr = ioutil.ReadAll(resp.Body)
		}
		if readErr != nil {
			return readErr
		}

		// 异常请求时设置状态码 StatusCode
		if resp.StatusCode != 200 {
			payload = []byte("StatusCode error:" + string(payload))
		}
		// 预读了数据，需要内容重新回写
		context.WithValue(ctx, "payload", payload)
		context.WithValue(ctx, "status_code", resp.StatusCode)

		resp.Body = ioutil.NopCloser(bytes.NewBuffer(payload))
		resp.ContentLength = int64(len(payload))
		resp.Header.Set("Content-Length", strconv.FormatInt(int64(len(payload)), 10))
		return nil
	}

	// 错误回调：当后台出现错误响应，会自动调用此函数
	// ModifyResponse 返回error，也会调用此函数
	// 为空时，出现错误返回502（错误网关）
	errFunc := func(w http.ResponseWriter, r *http.Request, err error) {
		http.Error(w, "ErrorHandler error:"+err.Error(), http.StatusInternalServerError)
	}

	return &httputil.ReverseProxy{
		Director:       director,
		Transport:      transport,
		ModifyResponse: modifyFunc,
		ErrorHandler:   errFunc}
}

// 重写NewSingleHostReverseProxy： 手动增加URL重写、更改内容、错误信息回调、连接池
func NewSingleHostReverseProxy(target *url.URL) *httputil.ReverseProxy {
	targetQuery := target.RawQuery // 编码查询值，不包含'?'
	// 定义director入口函数「管理者」
	director := func(req *http.Request) { // http://127.0.0.1:8080/realserver?a=1&b=2#a
		// 一个完整的url包含scheme、host、path、rawquery
		// scheme:http
		// host:127.0.0.1:8080
		// path:/realserver
		// rawquery:查询参数 a=1&b=2
		// target是下游服务器的信息，将其copy至req中进行重写
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = joinURLPath(target.Path, req.URL.Path)
		// 将上游客户端请求参数与下游请求参数进行合并
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	}

	// 重写修改返回响应内容
	modifyResponse := func(res *http.Response) error {
		fmt.Println("Here is modifyResponse function.")
		if res.StatusCode == http.StatusOK {
			srcBody, err := io.ReadAll(res.Body)
			if err != nil {
				fmt.Println("error when modifying Response!")
				return err
			}
			newBody := []byte(string(srcBody) + " Fxxk you!")
			res.Body = io.NopCloser(bytes.NewBuffer(newBody))
			length := int64(len(newBody))
			res.ContentLength = length
			res.Header.Set("Content-Length", strconv.Itoa(int(res.ContentLength)))
		}
		return nil
	}

	// 重写错误信息回调 当后台出现错误响应，会自动调用此函数
	// 为空时，出现错误返回502「错误网关」
	errorFunc := func(w http.ResponseWriter, r *http.Request, e error) {
		fmt.Println("Here is errHandler function.")
		http.Error(w, "ErrorHandler error:"+e.Error(), http.StatusInternalServerError)
	}

	// 连接池支持
	var transport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second, // 连接超时：拨号超时时间
			KeepAlive: 30 * time.Second, // 长连接超时时间
		}).DialContext,
		ForceAttemptHTTP2:     true,             // 是否强制http2
		MaxIdleConns:          100,              // 最大空闲连接数量
		IdleConnTimeout:       90 * time.Second, // 空闲连接超时时间
		TLSHandshakeTimeout:   10 * time.Second, // tls握手超时时间
		ExpectContinueTimeout: 1 * time.Second,  // 100-continue
	}

	return &httputil.ReverseProxy{
		Director:       director,
		ModifyResponse: modifyResponse,
		ErrorHandler:   errorFunc,
		Transport:      transport,
	}
}

// 合并a和b
// a : "" or "/"
// b : "/realserver"
func joinURLPath(a, b string) string {
	// a 后缀 和b前缀是否有斜杠
	aSlash := strings.HasSuffix(a, "/")
	bSlash := strings.HasPrefix(b, "/")

	switch {
	case aSlash && bSlash: // 保留a，去掉b
		return a + b[1:]
	case aSlash || bSlash: // 直接拼接
		return a + b
	}

	return a + "/" + b // 加/
}
