package main

import (
	"MSG/proxy/grpc_proxy/grpc_server_client/proto"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"io"
	"log"
	"time"
)

var msg = "this is client"

func main() {
	grpcConn, err := grpc.Dial("127.0.0.1:8085", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("Dial err : ", err)
		return
	}
	defer grpcConn.Close()

	grpcClient := proto.NewEchoClient(grpcConn)

	//UnaryEchoWithMetadata(grpcClient, msg)
	//ServerStreamWithMetadata(grpcClient, msg)
	//ClientStreamWithMetadata(grpcClient, msg)
	BidirectionalStreamWithMetadata(grpcClient, msg)
}

func UnaryEchoWithMetadata(c proto.EchoClient, msg string) {
	// 封装元数据
	md := metadata.Pairs("timestamp", time.Now().Format(time.StampNano)) // 两个参数为一组
	md.Append("hobby", "fxxk", "make", "love")
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	reply, err := c.UnaryEcho(ctx, &proto.EchoRequest{Message: msg})

	if err != nil {
		fmt.Printf("Dial err : ", err)
		return
	}
	fmt.Println("收到回复" + reply.String())
}

func ServerStreamWithMetadata(c proto.EchoClient, msg string) {
	// 封装元数据
	md := metadata.Pairs("timestamp", time.Now().Format(time.StampNano)) // 两个参数为一组
	md.Append("hobby", "fxxk", "make", "love")
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	ServerStreamingEchoClient, err := c.ServerStreamingEcho(ctx, &proto.EchoRequest{Message: msg})
	if err != nil {
		fmt.Printf("ServerStreamWithMetadata err : ", err)
		return
	}

	var rpcError error
	for {
		// 读取到流末尾：err == io.EOF
		reply, err := ServerStreamingEchoClient.Recv()
		if err != nil {
			if err == io.EOF {
				rpcError = err
				break
			}

		}
		fmt.Println("收到回复" + reply.String())
	}
	if rpcError != io.EOF {
		log.Fatalf("failed to finish ServerStreaming : %v", rpcError)
	}
}

func ClientStreamWithMetadata(c proto.EchoClient, msg string) {
	// 封装元数据
	md := metadata.Pairs("timestamp", time.Now().Format(time.StampNano)) // 两个参数为一组
	md.Append("hobby", "fxxk", "make", "love")
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	stream, err := c.ClientStreamingEcho(ctx)
	if err != nil {
		return
	}
	for i := 0; i < 10; i++ {
		err := stream.Send(&proto.EchoRequest{Message: msg})
		if err != nil {
			log.Fatalf("Send Failed!", err.Error())
		}
	}
	response, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("failed to finish ClientStream:%v", err)
	}
	fmt.Println("收到最后的回复" + response.String())
}

func BidirectionalStreamWithMetadata(c proto.EchoClient, msg string) {
	// 封装元数据
	md := metadata.Pairs("timestamp", time.Now().Format(time.StampNano)) // 两个参数为一组
	md.Append("hobby", "fxxk", "make", "love")
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	stream, err := c.BidirectionalStreamingEcho(ctx)
	if err != nil {
		return
	}
	// 开启两个协程
	go func() {
		for i := 0; i < 5; i++ {
			err := stream.Send(&proto.EchoRequest{Message: msg})
			if err != nil {
				log.Fatalf("Send Failed!", err.Error())
			}
		}
		// 可以不关闭 流「因为实际业务可能是动态生成的」
	}()

	var rpcError error
	for {
		// 读取到流末尾：err == io.EOF
		reply, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				rpcError = err
				break
			}

		}
		fmt.Println("收到回复" + reply.String())
	}
	if rpcError != io.EOF {
		log.Fatalf("failed to finish ServerStreaming : %v", rpcError)
	}
}
