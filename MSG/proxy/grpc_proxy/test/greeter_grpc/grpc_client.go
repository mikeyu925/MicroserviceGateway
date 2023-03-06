package main

import (
	"MSG/proxy/grpc_proxy/test/greeter_grpc/pb"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// 使用 gRPC 链接服务器
	// 抑制安全策略，不使用TLS层安全握手
	grpcConn, err := grpc.Dial("127.0.0.1:8004", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("Dial err : ", err)
		return
	}
	defer grpcConn.Close()
	// 初始化客户端
	grpcClient := pb.NewHelloServiceClient(grpcConn)
	reply, err := grpcClient.Hello(context.Background(), &pb.Person{Name: "小鱼", Age: 10})
	if err != nil {
		fmt.Printf("Dial err : ", err)
		return
	}
	fmt.Println("收到回复" + reply.String())
}
