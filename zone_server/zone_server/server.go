package zone_server

import (
	"fmt"
	"sync"

	"github.com/wqf/common_lib/codec"
	"github.com/wqf/common_lib/libnet"
	"github.com/wqf/zone_server/conf"
	"github.com/wqf/zone_server/zone_client"
)

type Server struct {
	sync.Mutex
	Server *libnet.Server
}

func init() {

}

func New() (s *Server) {
	s = new(Server)
	return
}

func (s *Server) Run() {
	protobuf := codec.Protobuf(Unpack, Packet)
	var err error
	s.Server, err = libnet.Serve("tcp", conf.Conf.TCPBind, protobuf, 0)
	if err != nil {
		panic(err)
	}
	fmt.Printf("start listeningL:%s \n", conf.Conf.TCPBind)
	go s.loop()
}

func (s *Server) loop() {
	for {
		session, err := s.Server.Accept() //接收客户端连接
		if err != nil {
			//记录错误
			continue
		}

		//session 处理
		client := zone_client.New(session) //因为库已经帮助持有了对象
		go client.ClientLoop()
	}
}
