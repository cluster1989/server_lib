package main

import (
	"flag"
	"time"

	"github.com/wqf/common_lib/libnet"
	"github.com/wqf/common_lib/perf"
	"github.com/wqf/common_lib/signal"
	"github.com/wqf/zone_server/conf"
)

func main() {

	//解析命令行
	flag.Parse()
	loadConfg()
	options := &libnet.ServerOptions{}
	options.Network = "tcp"
	options.Address = "127.0.0.1:6868"
	options.IsLittleIndian = false
	options.SendQueueBuf = 10
	options.RecvQueueBuf = 10

	options.SendTimeOut = time.Duration(180) * time.Second
	options.RecvTimeOut = time.Duration(180) * time.Second
	options.HeartBeatTime = time.Duration(60) * time.Second
	options.ReadTimeOutTimes = 3

	server := libnet.Serve(options)
	server.RegistRoute(100, func(content []byte, wildMsg bool) (args []interface{}) {
		args = make([]interface{}, 2)
		args = append(args, 1000)
		args = append(args, []byte{11, 22, 33, 44})
		return args
	}) //RegistHeartBeat
	server.RegistHeartBeat(102, func(content []byte, wildMsg bool) (args []interface{}) {
		args = make([]interface{}, 2)
		args = append(args, 1002)
		args = append(args, []byte{22, 33, 44, 55})
		return args
	})
	server.Run()
	perf.Init(conf.Conf.PprofBind)
	signal.InitSignal()
}

//解析配置文件
func loadConfg() {
	err := conf.InitConf()
	if err != nil {
		panic(err)
	}
}
