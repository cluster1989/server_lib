package main

import (
	"flag"

	"github.com/wqf/common_lib/perf"
	"github.com/wqf/common_lib/signal"
	"github.com/wqf/zone_server/conf"
	"github.com/wqf/zone_server/zone_route"
	"github.com/wqf/zone_server/zone_server"
)

func main() {

	//解析命令行
	flag.Parse()
	loadConfg()
	server := zone_server.New()
	server.Run()
	zone_route.RegisterAllRoute()
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
