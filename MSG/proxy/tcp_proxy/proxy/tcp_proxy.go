package tcp_proxy

import (
	"context"
	"io"
	"log"
	"net"
	"time"
)

type TCPReverseProxy struct {
	Addr string          // 下游真实服务器地址 host:port
	Ctx  context.Context // 上下文

	DialTimeout     time.Duration // 拨号超时时间 持续时间
	Deadline        time.Duration // 拨号截止时间 截止时间
	KeepAlivePeriod time.Duration // 长连接超时时间

	// 拨号器 支持自定义：拨号成功，返回连接；拨号失败，返回error
	DialContext func(ctx context.Context, network, address string) (net.Conn, error)

	// TCP 整合负载均衡算法
	Director func(remoteAddr string) (string, error)

	// 修改响应  从连接里拿数据 「如果返回错误，则由ErrHandler处理」
	ModifyResponse func(conn net.Conn) error
	// 错误处理
	ErrorHandler func(conn net.Conn, e error)
}

func NewSingleHostReverseProxy(addr string) *TCPReverseProxy {
	if addr == "" {
		panic("TCP must not be empty!")
	}
	return &TCPReverseProxy{
		Addr:            addr,
		DialTimeout:     10 * time.Second,
		Deadline:        60 * time.Second,
		KeepAlivePeriod: time.Hour,
	}
}

/*
完成上下游连接，及数据的交换
接收上游连接->向下游发送请求->接收下游响应->拷贝/修改，响应至上游
*/
func (tpxy *TCPReverseProxy) ServerTCP(src net.Conn, ctx context.Context) {
	var cancel context.CancelFunc
	if tpxy.DialTimeout >= 0 { // 连接超时时间
		ctx, cancel = context.WithTimeout(ctx, tpxy.DialTimeout)
	}
	if tpxy.Deadline >= 0 { // 连接截止时间
		ctx, cancel = context.WithDeadline(ctx, time.Now().Add(tpxy.Deadline))
	}
	if cancel != nil {
		defer cancel()
	}
	// 如果没有自定义拨号器，则使用系统默认的拨号器
	if tpxy.DialContext == nil {
		tpxy.DialContext = (&net.Dialer{
			Timeout:   tpxy.DialTimeout,
			KeepAlive: tpxy.KeepAlivePeriod,
		}).DialContext
	}

	// 执行入口函数
	tpxy.Director(src.RemoteAddr().String())

	// 向下游发送请求
	dst, err := tpxy.DialContext(ctx, "tcp", tpxy.Addr)
	if err != nil {
		// TODO 错误处理
		tpxy.getErrHandler()(src, err)
		src.Close()
		return
	}
	defer dst.Close() // 关闭下游连接
	// 修改下游服务器响应
	if !tpxy.modifyResponse(dst) {
		return
	}
	// 从下游拷贝至上游 // modify

	if _, err = tpxy.bytesCopy(src, dst); err != nil {
		// TODO 错误处理
		tpxy.getErrHandler()(dst, err)
		dst.Close()
		return
	}
}

// 通过此函数修改响应，如果没有问题，则返回true；否则返回false
func (pxy *TCPReverseProxy) modifyResponse(res net.Conn) bool {
	if pxy.ModifyResponse == nil {
		return true
	}
	if err := pxy.ModifyResponse(res); err != nil {
		res.Close() // 关闭连接
		pxy.getErrHandler()(res, err)
		return false
	}
	return true
}

func (pxy *TCPReverseProxy) bytesCopy(dst, src net.Conn) (len int64, err error) {
	return io.Copy(dst, src)
}

func (pxy *TCPReverseProxy) getErrHandler() func(conn net.Conn, e error) {
	if pxy.ErrorHandler == nil {
		return pxy.defaultErrorHandler
	}
	return pxy.ErrorHandler
}

func (pxy *TCPReverseProxy) defaultErrorHandler(conn net.Conn, e error) {
	log.Printf("TCP conn '%v' error : %v. ", conn.RemoteAddr().String(), e)
}
