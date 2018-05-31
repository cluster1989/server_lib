package libnet2

import "time"

// 连接的配置
type NetOption struct {
	// 最大的连接数目
	MaxConn int32

	// 网络类型
	Network string

	// 连接地址和端口
	Address string

	// 多少个核心
	Workers int
}

// session 的配置
type SessionOption2 struct {
	ReadTimeout time.Duration //读取的超时
	// 允许的超时次数
	ReadTimeoutTimes int

	RecvChanSize int //接收和发送队列的大小
	SendChanSize int
}
