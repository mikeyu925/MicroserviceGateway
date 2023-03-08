package proxy

import (
	"MSG/loadbalance"
	"MSG/proxy/grpc_proxy"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

func NewGrpcLoadBalanceReverseProxy(lb loadbalance.LoadBalance) grpc.StreamHandler {
	return func() grpc.StreamHandler {
		director := func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
			nextAddr, err := lb.Get(fullMethodName)
			if err != nil {
				log.Fatal("Get next address fail!")
			}
			conn, err := grpc.DialContext(ctx, nextAddr,
				// 禁用安全传输
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			return ctx, conn, err
		}
		return grpc_proxy.TransparentHandler(director)
	}()
}
