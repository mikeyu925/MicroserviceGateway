package proxy

import (
	"MSG/loadbalance"
	tcp_proxy "MSG/proxy/tcp_proxy/proxy"
	"context"
	"log"
	"time"
)

// 1. 选择合适的负载均衡器
// 2. 创建一个支持负载均衡算法的handler
// 3. 启动TCP代理服务
func NewTcpLoadBalanceReverseProxy(c context.Context, lb loadbalance.LoadBalance) *tcp_proxy.TCPReverseProxy {
	pxy := &tcp_proxy.TCPReverseProxy{
		Ctx:             c,
		Deadline:        time.Minute,
		DialTimeout:     10 * time.Second,
		KeepAlivePeriod: time.Hour,
	}
	pxy.Director = func(remoteAddr string) (nextAddr string, err error) {
		nextAddr, err = lb.Get(remoteAddr)
		if err != nil {
			log.Fatal("Get next address error!")
		}
		pxy.Addr = nextAddr
		return
	}
	return pxy
}
