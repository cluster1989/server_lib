package main

import (
	"context"
	"fmt"

	"github.com/wuqifei/chat/common/rpc_model"
	"github.com/wuqifei/server_lib/libgrpc"
	"google.golang.org/grpc/grpclog"
)

func main() {

	options := &libgrpc.ClientOptions{}
	options.Address = "127.0.0.1:8124"
	client := libgrpc.NewClient(options)
	// 初始化客户端

	in := &rpc_model.LogicServerModel_ReqRegister{}
	in.Addr = "127.0.0.1:8123"
	in.Password = "123456789"
	in.UserName = "wqfwqf"

	l := rpc_model.NewLogicServerRPCClient(client.ClientConn)
	res, err := l.Register(context.Background(), in)

	// res, err := c.SayHello(context.Background(), req)

	if err != nil {
		grpclog.Fatalln(err)
	}
	fmt.Println(res.Uid)
	state := client.ClientConn.GetState()
}
