package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type TCPHandler interface {
	// 提供Tcp服务的函数
	ServerTCP(conn net.Conn, ctx context.Context)
}

type contextKey struct {
	name string
}

var (
	ErrServerClosed     = errors.New("tcp: Server closed")
	ErrAbortHandler     = errors.New("net/tcp: abort Handler")
	ServerContextKey    = &contextKey{"tcp-server"}
	LocalAddrContextKey = &contextKey{"local-addr"} // 当前本机地址
)

// 原生默认的handler
type tcpHandler struct {
}

func (t *tcpHandler) ServerTCP(conn net.Conn, ctx context.Context) {
	// 接收数据
	buffer := make([]byte, 1024)
	n, _ := conn.Read(buffer)
	fmt.Println("tcpHandler 收到了一个请求:", string(buffer[:n]))
	conn.Write([]byte("我是tcpHandler! 我们收到了你的TCP请求！\n"))
}

// TCP核心结构体，监听主机，并提供服务
// Addr 主机地址
// Handler 回调函数，处理TCP请求
type TCPServer struct {
	Addr        string
	Handler     TCPHandler
	BaseContext context.Context // 上下文实例，收集取消、终止、错误等信息
	err         error           // TCP error

	ReadTimeOut      time.Duration // 读超时
	WriteTimeOut     time.Duration // 写超时
	KeepAliveTimeout time.Duration // 长连接超时

	mu         sync.Mutex         // 连接关闭等关键动作时用到锁
	doneChan   chan struct{}      // 服务已完成，监听系统信号
	inShutDown int32              // 服务终止：0 未关闭 1 已关闭
	l          *onceCloseListener // 服务器监听器，使用完成时的动作
}

func (srv *TCPServer) ListenAndServe() error {
	if srv.shuttingDown() {
		return ErrServerClosed
	}
	addr := srv.Addr
	if addr == "" {
		return errors.New("we need an Address!")
	}
	if srv.Handler == nil {
		srv.Handler = &tcpHandler{}
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return srv.Serve(ln)
}

func ListenAndServe(addr string, handler TCPHandler) error {
	server := &TCPServer{
		Addr:    addr,
		Handler: handler,
	}
	return server.ListenAndServe()
}

// 提供服务
func (srv *TCPServer) Serve(l net.Listener) error {

	srv.l = &onceCloseListener{Listener: l}
	defer l.Close()
	// 获取当前默认上下文
	if srv.BaseContext == nil {
		srv.BaseContext = context.Background()
	}
	baseCtx := srv.BaseContext

	ctx := context.WithValue(baseCtx, ServerContextKey, srv)
	for {
		// 尝试获取连接
		rw, err := l.Accept()
		if err != nil {
			select {
			case <-srv.getDoneChan():
				return ErrServerClosed
			default:
			}
			return err
		}

		c := srv.newConn(rw) // 对tcp连接的二次封装
		go c.serve(ctx)      // 启动一个协程提供服务
	}
}

func (s *TCPServer) getDoneChan() <-chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.doneChan == nil {
		s.doneChan = make(chan struct{})
	}
	return s.doneChan
}
func (srv *TCPServer) Close() error {
	// 使用原子操作
	atomic.StoreInt32(&srv.inShutDown, 1)
	close(srv.doneChan)
	srv.l.Close()
	return nil
}

func (s *TCPServer) shuttingDown() bool {
	return atomic.LoadInt32(&s.inShutDown) == 1
}

type conn struct {
	server     *TCPServer
	rwc        net.Conn // 连接
	remoteAddr string
}

// Create new connection from rwc.
func (srv *TCPServer) newConn(rwc net.Conn) *conn {
	c := &conn{
		server:     srv,
		rwc:        rwc,
		remoteAddr: rwc.RemoteAddr().String(),
	}
	// 设置参数：从TCPServer 中取字段，赋值给TCPConn
	if t := srv.ReadTimeOut; t != 0 {
		c.rwc.SetReadDeadline(time.Now().Add(t)) // 设置读截止时间
	}
	if t := srv.WriteTimeOut; t != 0 {
		c.rwc.SetWriteDeadline(time.Now().Add(t)) // 设置写截止时间
	}
	if t := srv.KeepAliveTimeout; t != 0 {
		if tcpConn, ok := c.rwc.(*net.TCPConn); ok {
			tcpConn.SetKeepAlive(true)
			tcpConn.SetKeepAlivePeriod(t)
		}
	}
	return c
}

func (c *conn) serve(ctx context.Context) {
	defer func() {
		if err := recover(); err != nil && err != ErrAbortHandler {
			const size = 64 << 10 // 65536
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			fmt.Printf("http: panic serving %v: %v\n%s", c.remoteAddr, err, buf)
		}
		// 服务完之后关闭连接
		c.rwc.Close()
	}()
	ctx = context.WithValue(ctx, LocalAddrContextKey, c.rwc.LocalAddr())
	if c.server.Handler == nil {
		panic("TCP Handler is nil!")
	}
	c.server.Handler.ServerTCP(c.rwc, ctx)
}

type onceCloseListener struct {
	net.Listener
	once     sync.Once //只执行一次
	closeErr error
}

func (oc *onceCloseListener) Close() error {
	oc.once.Do(oc.close)
	return oc.closeErr

}

func (oc *onceCloseListener) close() { oc.closeErr = oc.Listener.Close() }
