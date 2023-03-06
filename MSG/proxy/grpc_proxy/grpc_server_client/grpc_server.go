package main

import (
	"MSG/proxy/grpc_proxy/grpc_server_client/proto"
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"io"
	"log"
	"net"
)

var port = flag.Int("port", 8005, "the port to serve on")

func main() {
	flag.Parse()
	s := grpc.NewServer()
	proto.RegisterEchoServer(s, &MyServer{})

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed listening: %v", err)
	}
	s.Serve(listener)
}

type MyServer struct {
}

func (ms *MyServer) UnaryEcho(ctx context.Context, req *proto.EchoRequest) (*proto.EchoResponse, error) {
	fmt.Println("------一元服务器端被调用-------")

	//type MD map[string][]string
	// 从上下文拿到元数据
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		// 根据需要进行解析
		fmt.Println("md : ", md)
	} else {
		log.Println("miss metadata from context")
	}

	return &proto.EchoResponse{Message: req.Message}, nil
}

// ServerStreamingEcho is server side streaming.
func (ms *MyServer) ServerStreamingEcho(req *proto.EchoRequest, stream proto.Echo_ServerStreamingEchoServer) error {
	fmt.Println("------服务端流式处理-------")
	for i := 0; i < 5; i++ {
		err := stream.Send(&proto.EchoResponse{Message: req.Message})
		if err != nil {
			return err
		}
	}
	return nil
}

// ClientStreamingEcho is client side streaming.
func (ms *MyServer) ClientStreamingEcho(stream proto.Echo_ClientStreamingEchoServer) error {
	fmt.Println("------客户端流式处理-------")
	var Errinfo error
	for {
		req, err := stream.Recv() // 返回单个的请求
		if err != nil {
			Errinfo = err
			break
		}
		fmt.Println("服务器接收到信息: " + req.String())
	}
	if Errinfo != io.EOF {
		return Errinfo
	}
	fmt.Println("成功接收到所有消息!")
	return stream.SendAndClose(&proto.EchoResponse{Message: "Get All MSG!"}) // 发送消息并且关闭

}

// BidirectionalStreamingEcho is bidi streaming.
func (ms *MyServer) BidirectionalStreamingEcho(stream proto.Echo_BidirectionalStreamingEchoServer) error {
	fmt.Println("------双向流式处理-------")
	// 所谓的双向流，发送接收是并行的「接收一条处理一条」
	var Errinfo error
	for {
		req, err := stream.Recv() // 返回单个的请求
		if err != nil {
			Errinfo = err
			break
		}
		fmt.Println("服务器接收到信息: " + req.String())
		// 也可以开启一个线程去发送
		err = stream.Send(&proto.EchoResponse{Message: "okok!"})
		if err != nil {
			return err
		}
	}
	if Errinfo != io.EOF {
		return Errinfo
	}
	fmt.Println("成功接收到所有消息!")
	return stream.Send(&proto.EchoResponse{Message: "Get All MSG!"}) // 发送消息并且关闭

}
