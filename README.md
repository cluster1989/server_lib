# 简单的TCP服务框架

测试代码，可以参照example里面的文件，性能测试，做了一个简单的，里面由参考的截图，1000并发下的测试。

使用类似于http框架的使用
设置router，然后只需要关心router之间的逻辑，不需要关心对于连接的处理

内部使用了非阻塞的通道来接发数据

里面包含了ini类似的配置文件的库，可以用这个库，来配置自己的配置文件，详情可以参考libconf文件夹

去除代码中的用户管理模块，时间轮模块，保持代码功能单一性

第三方库：

    github.com/garyburd/redigo/redis    

如果用到redis的lib就可以用这个，不用的话，可以删除，就没有第三方库了

接口说明：

```
type ServerOptions struct {
	// 类型
	Network string

	// 地址
	Address string

	// cpu大小头设置
	IsLittleIndian bool

	// 接收和发送的队列大小
	SendQueueBuf int
	RecvQueueBuf int

	//接收和发送的超时时间
	SendTimeOut time.Duration
	RecvTimeOut time.Duration

	// 允许超时次数
	ReadTimeOutTimes int

	// 最大的接收字节数
	MaxRecvBufferSize int

	// 最大的发送字节数
	MaxSendBufferSize int
}
```

上面这个是配置的配置参数，最好自己定义一下，并没有最佳的参数配置。可以按照自己的物理机器测试。

// 初始化一个tcp的服务

```
server := libnet.Serve(options)
```

// 注册服务
// 这里需要注意的是，心跳的注册，和一般的注册，最好分开，因为我无法区分哪个是心跳，所以分别作出了两个接口

```
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
```

// 启动服务

```
go server.Run()
```
