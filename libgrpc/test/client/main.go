package main

import (
	"context"
	"fmt"

	"github.com/wuqifei/server_lib/signal"
	"github.com/wuqifei/test/libgrpc"
	"github.com/wuqifei/test/libgrpc/test/pb"
	"google.golang.org/grpc/grpclog"
)

func main() {

	options := &libgrpc.ClientOptions{}
	options.Address = "127.0.0.1:9999"
	client := libgrpc.NewClient(options)
	// 初始化客户端
	c := pb.NewHelloClient(client.ClientConn)

	// 调用方法
	req := &pb.HelloRequest{Name: "gRPC"}
	res, err := c.SayHello(context.Background(), req)

	if err != nil {
		grpclog.Fatalln(err)
	}
	fmt.Println(res.Message)
	signal.InitSignal()
}
