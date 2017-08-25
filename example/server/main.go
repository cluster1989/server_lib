package main

import (
	"flag"
	"time"

	"github.com/wuqifei/server_lib/libnet"
	"github.com/wuqifei/server_lib/logs"
	"github.com/wuqifei/server_lib/signal"
)

func main() {
	libServer()
	signal.InitSignal()
}

func libServer() {
	//解析命令行
	flag.Parse()
	logger := logs.GetLibLogger()
	logger.SetLogger("console", `{"color":false}`)
	logger.EnableFuncCallDepth(true)

	options := &libnet.ServerOptions{}
	options.Network = "tcp"
	options.Address = "127.0.0.1:6868"
	options.IsLittleIndian = false
	options.SendQueueBuf = 10
	options.RecvQueueBuf = 10

	options.SendTimeOut = time.Duration(180) * time.Second
	options.RecvTimeOut = time.Duration(180) * time.Second //5s 超时间
	options.HeartBeatTime = time.Duration(60) * time.Second
	options.ReadTimeOutTimes = 3
	options.MaxRecvBufferSize = 8
	options.MaxSendBufferSize = 8

	server := libnet.Serve(options)
	server.RegistRoute(100, func(content []byte, wildMsg bool) (args []interface{}) {
		args = make([]interface{}, 0)
		args = append(args, uint16(1000))
		args = append(args, []byte{11, 22, 33, 44})
		return args
	}) //RegistHeartBeat
	server.RegistHeartBeat(102, func(content []byte, wildMsg bool) (args []interface{}) {
		args = make([]interface{}, 0)
		args = append(args, uint16(1002))
		args = append(args, []byte{22, 33, 44, 55})
		return args
	})
	go server.Run()
}
