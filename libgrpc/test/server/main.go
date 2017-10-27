package main

import (
	"context"
	"fmt"

	"github.com/wuqifei/server_lib/libgrpc"
	"github.com/wuqifei/server_lib/libgrpc/test/pb"
	"github.com/wuqifei/server_lib/signal"
)

// 定义helloService并实现约定的接口
type helloService struct{}

// HelloService Hello服务
var HelloService = helloService{}

// SayHello 实现Hello服务接口
func (h helloService) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloResponse, error) {
	resp := new(pb.HelloResponse)
	resp.Message = fmt.Sprintf("Hello %s.", in.Name)

	return resp, nil
}

func main() {
	options := &libgrpc.ServerOptions{}
	options.Address = "127.0.0.1:9999"
	server := libgrpc.NewServer(options)
	pb.RegisterHelloServer(server.Server, HelloService)
	go server.RPCServe()
	signal.InitSignal()
}
