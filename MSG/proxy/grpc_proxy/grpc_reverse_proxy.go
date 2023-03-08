package grpc_proxy

import (
	"MSG/proxy/grpc_proxy/grpc_server_client/proto"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io"
	"strings"
)

type handler struct {
	director StreamDirector
}

// 构建一个下游连接器：ClientStream
// 创建下游连接：往下游真实服务器创建连接
// 上游和下游数据拷贝
// 关闭双向流
func (h *handler) handler(srv interface{}, proxyServerStream grpc.ServerStream) error {
	fmt.Println("收到了请求！正在经过代理服务器...")
	// 过滤非RPC请求
	// 得到 "/service/method
	method, ok := grpc.MethodFromServerStream(proxyServerStream)
	if !ok { // 非rpc请求
		return status.Errorf(codes.Internal, "There is no rpc-request!")
	}
	// 不处理内部请求
	if strings.HasPrefix(method, "/com.example.internal") {
		return status.Errorf(codes.Unimplemented, "Unimplemented Method!")
	}

	// 构建下游连接器
	ctx := proxyServerStream.Context()
	// 负载均衡算法获取下游服务器地址
	ctx, pxyclientConn, err := h.director(ctx, method)
	if err != nil {
		return err
	}
	defer pxyclientConn.Close()

	md, _ := metadata.FromIncomingContext(ctx)       // 从上游请求上下文得到元数据
	outCtx, clientCancel := context.WithCancel(ctx)  // 获取取消函数「可能为nil」
	outCtx = metadata.NewOutgoingContext(outCtx, md) // 封装一个新的下游请求context

	// 封装下游客户端流实例
	pxyStreamDesc := &grpc.StreamDesc{
		ServerStreams: true,
		ClientStreams: true,
	}
	// 代理的下游客户端流
	pxyclientStream, err := grpc.NewClientStream(outCtx, pxyStreamDesc, pxyclientConn, method)
	if err != nil {
		return err
	}
	// 上游与下游数据拷贝「应该是并行的」，两个协程之间通过管道进行通信
	chanc2s := h.clientToServer(pxyclientStream, proxyServerStream) // 上游请求消息发送给下游真实服务器
	chans2c := h.serverToClient(proxyServerStream, pxyclientStream) // 下游服务器发送响应给上游的客户端

	// 关闭双向流
	// C/S双方谁会先关闭channel，是不确定的，因此用select语句进行随机选择
	for i := 0; i < 2; i++ { // 因为两个都要执行到一遍，但是不保证顺序
		select {
		case s2cErr := <-chans2c: // 向上游回写消息
			// Trailer:metadata，当流被关闭「ClientStream」，读取消息得到 error，都会生成元数据
			proxyServerStream.SetTrailer(pxyclientStream.Trailer()) // 写回Trailer
			if s2cErr != io.EOF {                                   // 出现错误
				// proxyServerStream 不需要关闭
				return s2cErr
			}
			return nil
		case c2sErr := <-chanc2s: // 向下游发送请求
			if c2sErr == io.EOF {
				// 接收到了发送结束的信号，并且不再发送
				// 关闭代理客户端发送流
				pxyclientStream.CloseSend()
			} else {
				// 发送过程中出现了问题
				// 取消发送，并返回错误
				if clientCancel != nil {
					clientCancel()
				}
				return status.Errorf(codes.Internal, "Failed client2server : %v\n", c2sErr)
			}
		}
	}
	return nil
}

// 上游客户端流向服务端 == 向下游发送消息
func (h *handler) clientToServer(dst grpc.ClientStream, src grpc.ServerStream) chan error {
	res := make(chan error, 1)
	// 开启一个协程
	go func() {
		msg := &proto.EchoRequest{}
		for {
			// 请求头没有必要设置
			// 服务器只有读取到第一条客户消息的同时，才可以读取请求头
			if err := src.RecvMsg(msg); err != nil { // RecvMsg是阻塞的
				res <- err
				break
			}
			err := dst.SendMsg(msg)
			if err != nil {
				res <- err
				break
			}
		}
	}()
	return res
}

// 下游服务端响应流向客户端 == 向上游响应消息
func (h *handler) serverToClient(dst grpc.ServerStream, src grpc.ClientStream) chan error {
	res := make(chan error, 1)
	// 开启一个协程
	go func() {
		msg := &proto.EchoResponse{}
		for i := 0; ; i++ {
			// 对response的header进行处理，因为客户端是先读服务器的响应头，然后再做出响应的一系列处理,所以要设置响应头
			if i == 0 { // 仅仅第一次设置响应头
				md, err := src.Header() // 获取响应头
				if err != nil {
					res <- err
					break
				}
				// 设置响应头
				err = dst.SetHeader(md)
				if err != nil {
					res <- err
					break
				}
			}
			if err := src.RecvMsg(msg); err != nil { // RecvMsg是阻塞的
				res <- err // 可能是io.EOF
				break
			}
			err := dst.SendMsg(msg)
			if err != nil {
				res <- err
				break
			}
		}
	}()
	return res
}

type StreamDirector func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error)

func TransparentHandler(director StreamDirector) grpc.StreamHandler {
	streamer := &handler{director: director}
	return streamer.handler
}
