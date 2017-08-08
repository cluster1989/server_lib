package conf

import (
	"flag"
	"time"

	"github.com/wqf/common_lib/conf_tool"
)

var (
	confUtil *conf_tool.Config
	confFile string //config 文件目录

	//Conf 暴露出来的config对象
	Conf *Config
)

// Config 配置文件对象
type Config struct {
	//base confg
	PidFile   string   `soulte:"base:pidfile"`
	PprofBind []string `soulte:"base:pprof.bind:,"`
	Debug     bool     `soulte:"base:debug"`
	// tcp
	TCPBind   string `soulte:"tcp:bind"`
	TCPSndBuf int    `soulte:"tcp:sndbuf:memory"`
	TCPRcvBuf int    `soulte:"tcp:rcvbuf:memory"`
	// 消息的ack等，会有个超时的限制，超过时间之后，服务器会主动断开
	WriteTimeout time.Duration `soulte:"tcp:writeTimeout:time"`
	ReadTimeout  time.Duration `soulte:"tcp:readTimeout:time"`
	//rpc setting
	RPCServerAddr     string        `soulte:"rpc:rpc.addr"`
	RPCZKServers      []string      `soulte:"rpc:rpc.zkaddrs:,"`
	RPCZKNode         string        `soulte:"rpc:rpc.zknode"`
	RPCUpdateInterval time.Duration `soulte:"rpc:rpc.rpcinterval:time"`
	RCPServerName     string        `soulte:"rpc:rpc.rpcname"`
	//redis
	DialReadTimeout  time.Duration `soulte:"redis:redis.readtimeout:time"`
	DialWriteTimeout time.Duration `soulte:"redis:redis.writetimeout:time"`
	RedisAddress     string        `soulte:"redis:redis.addr"`
	MaxIdle          int           `soulte:"redis:redis.maxidle"`
	MaxActive        int           `soulte:"redis:redis.maxactive"`
	//zookeeper
	ZKServers        []string      `soulte:"zookeeper:zookeeper.zkaddrs:,"`
	ZKNode           string        `soulte:"zookeeper:zookeeper.zknode"`
	ZKSessionTimeout time.Duration `soulte:"zookeeper:zookeeper.zksessiontimeout:time"`
	ZKDialTimeOut    time.Duration `soulte:"zookeeper:zookeeper.zkdialtimeout:time"`
}

func init() {
	flag.StringVar(&confFile, "c", "./zone_server/zone.conf", "set zone config path")
}

// InitConf 初始化配置文件
func InitConf() (err error) {

	conf := new(Config)
	confUtil = conf_tool.New()
	if err = confUtil.Parse(confFile); err != nil {
		return
	}
	if err = confUtil.Unmarshal(conf); err != nil {
		return
	}
	Conf = conf
	return
}

// ReloadConf 重载配置文件
func ReloadConf() (err error) {
	conf := new(Config)
	confUtil, err = confUtil.Reload() //重新装载
	if err != nil {
		return
	}
	if err = confUtil.Unmarshal(conf); err != nil {
		return
	}
	Conf = conf
	return
}
