package main

import (
	"flag"
	"time"

	"github.com/wuqifei/server_lib/libnet/libsession"
	"github.com/wuqifei/server_lib/perf"

	"os"
	"runtime/pprof"

	"github.com/wuqifei/server_lib/libnet"
	"github.com/wuqifei/server_lib/logs"
	"github.com/wuqifei/server_lib/signal"
)

func main() {
	initLogger()
	libServer()
	ips := []string{":8080"}
	perf.Init(ips)

	f, err := os.Create("cpu.prof")
	if err != nil {
		logs.Error(err)
	}
	pprof.StartCPUProfile(f)
	go func() {
		time.Sleep(time.Duration(60*20) * time.Second)
		pprof.StopCPUProfile()
	}()
	signal.InitSignal()
}

func initLogger() {

	logger := logs.GetLibLogger()
	logger.SetLogger("console", `{"color":false}`)
	logger.EnableFuncCallDepth(true)

}

func libServer() {
	//解析命令行
	flag.Parse()

	options := &libnet.ServerOptions{}
	options.Network = "tcp"
	options.Address = ":6868"
	options.SessionOption = libsession.Options{}
	options.SessionOption.IsLittleEndian = false
	options.SessionOption.SendChanSize = 10
	options.SessionOption.RecvChanSize = 10

	options.SessionOption.ReadTimeout = time.Duration(180) * time.Second //5s 超时间
	options.SessionOption.ReadTimeoutTimes = 3
	options.MaxRecvBufferSize = 8 * 1024
	options.MaxSendBufferSize = 8 * 1024

	server := libnet.Serve(options)
	server.RegistRoute(100, func(content []byte, sessionID uint64) (args []interface{}, err error) {
		args = make([]interface{}, 0)
		args = append(args, uint16(1000))
		args = append(args, []byte{11, 22, 33, 44})
		return args, nil
	})

	server.OnClose = OnClose

	go server.Run()
}
func OnClose(sessID uint64) {
	logs.Info("sessID:%l", sessID)
}
