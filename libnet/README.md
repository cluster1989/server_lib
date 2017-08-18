# lib net

说明
=====

tcp的一个微型框架

-----

对于数据处理
Protocol接口
需要实现NewCodec接口
返回Codec的一个实现
在这里可以让框架处理接收，发送，关闭请求

-----

tcp服务的使用

function

使用方式如下

启动tcp服务
libnet.Serve("tcp","localhost:8080",Codec实体类,通道大小)
发送的通道是否是同步，还是需要进行一定的队列处理

libnet.Connect("")
连接tcp用的库

libnet.NewServer("连接的listener","Codec的实体类",通道大小)

libnet.Listener()返回listener

libnet.Accept()
里面是一个循环，会一直去轮询接收客户端连接
如果遇到错误，会持续的去等待

在遇到连接之后会生成一个客户端session

libnet.Stop()
关闭连接，并且释放客户端连接


----------

session的使用
session的Extra可以用来挂载，所需要使用的额外的session信息

session启动之后，会启动一个session的发送的loop检测
所有的发送都会放入，发送的通道中，进行发送，一旦通道有数据，则第一时间发送除去


session.ID() 返回seessionid

session.Codec()返回Codec的实现类

session.Close() 在确认关闭之后，会调用关闭的回调

session.Recv() 会调用codec的接收方法

session.Send() 会调用发送的方法，但是不会直接调用，会将msg放入sendchan中，然后发送

session.AddCloseCallback(key,value) 增加session 关闭回调

session.RemoveCloseCallback(key) 删除session 关闭回调

简单使用如下

``` go
func main() {
	protobuf := codec.Protobuf()
	server, err := libnet.Serve("tcp", "127.0.0.1:8989", protobuf, 0)
	if err != nil {
		fmt.Printf("error:%v", err)
	}
	s = server
	loop()
	// signal.InitSignal()
}

func loop() {
	for {
		session, err := s.Accept()
		if err != nil {

			fmt.Printf("loop error:%v", err)
		}
		fmt.Printf("a new client:%ld \n\r", session.ID())
		go sessionLoop(session)
	}
}

func sessionLoop(session *libnet.Session) {
	for {
		data, err := session.Recv()
		fmt.Printf("receive data ---- :%v \n\r", data)

		if err != nil {

			fmt.Printf("sessionLoop error:%v \n\r", err)
		}
		if data != nil {

			q := test_proto.ReqLogin{}
			err := proto.Unmarshal(data, &q)
			if err == nil {

				fmt.Printf("data transfer:%v \n\r", &q)
			}
			s := &test_proto.ResLogin{
				Cmd:     1,
				CmdStr:  "12311212",
				ErrCode: 10,
				ErrStr:  "empty",
			}
			session.Send(s)
		}
	}
}
```