package main

import (
	"MSG/proxy/grpc_proxy/test/greeter_grpc/pb"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"net"
)

type HelloService struct {
}

func (hs *HelloService) Hello(c context.Context, p *pb.Person) (*pb.Person, error) {
	reply := &pb.Person{
		Name: "ywh" + p.Name,
		Age:  24 + p.Age,
	}
	return reply, nil
}

func main() {
	// 注册gRPC服务，绑定对象方法
	grpcServer := grpc.NewServer()
	pb.RegisterHelloServiceServer(grpcServer, &HelloService{})
	// 创建设置监听
	listener, err := net.Listen("tcp", "127.0.0.1:8004")
	if err != nil {
		fmt.Println("listen err!")
		return
	}
	fmt.Println("listen port 127.0.0.1:8004 ....")
	defer listener.Close()

	// 绑定服务：将监听绑定rpc服务
	grpcServer.Serve(listener)
}
