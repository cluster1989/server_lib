package libnet2

import (
	"net"

	"github.com/wuqifei/server_lib/libio"
)

// session收到信息
type OnSessRecv func(sess Session2Interface, val []byte)

// session关闭
type OnSessClose func(sess Session2Interface)

// session错误
type OnSessError func(sess Session2Interface, err error)

// 收到错误
type OnError func(err error)

// session关闭
type OnClose func(sess Session2Interface)

// 收到一个session对象
type OnSession func(sess Session2Interface)

// tcp的session对象
type Session2Interface interface {

	// 监听
	Accept()

	//  发送数据
	Send([]byte) error

	// 关闭
	Close() error

	// 收到信息
	Recv(OnSessRecv)

	// 设置参数
	Set(key, val interface{}) error

	// 获取参数
	Get(key interface{}) (interface{}, error)

	// 删除参数
	Del(key interface{}) (bool, error)

	// 清空参数
	Clear() error

	// 获取连接
	GetConn() net.Conn

	// 获取内存存储的uniquid
	GetUniqueID() uint64

	Reader() *libio.Reader
	Writer() *libio.Writer
}

//  包的策略
type PacketInterface interface {
	// 不要去猜测读写策略，每个人可能不一样，也不要约定读写策略，没必要
	Write(w *libio.Writer, b []byte) error
	Read(r *libio.Reader) ([]byte, error)
}

// 接口
type LibserverInterface interface {
	// 网络监听返回
	Listener() net.Listener

	// 启动
	Run()

	// 关闭
	Close()
}
