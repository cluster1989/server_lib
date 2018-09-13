package tcp_server

import (
	"fmt"
	"sync"
	"time"

	"github.com/wuqifei/server_lib/libio"
	"github.com/wuqifei/server_lib/libnet2"
)

var (
	server libnet2.LibserverInterface
)

func NewServer() {
	server, _ = libnet2.New()

	server.Run()
	libnet2.ServerSessionBlock = func(sess libnet2.Session2Interface) {
		fmt.Printf("sess [%d] \n", sess.GetUniqueID())
		sess.Send([]byte("hello new connect"))
	}

	libnet2.ServerErrorBlock = func(err error) {
		fmt.Printf("error of block [%v]\n", err)
	}

	libnet2.SessionCloseBlock = func(sess libnet2.Session2Interface) {
		fmt.Printf("session [%d]closed \n", sess.GetUniqueID())
	}

	libnet2.SessionErrorBlock = func(sess libnet2.Session2Interface, err error) {

		fmt.Printf("session [%d] error of block [%v]\n", sess.GetUniqueID(), err)
	}
	libnet2.ServerPacket = new(packet)

	// 收到信息
	libnet2.SessionRecvBlock = func(sess libnet2.Session2Interface, val []byte) {
		fmt.Printf("[%v]:server sess[%d] recv:[%s] \n", time.Now(), sess.GetUniqueID(), string(val))
		sess.Send([]byte("hello there"))
		once := sync.Once{}
		once.Do(func() {

			server.UpdateSessionID(1, 2)
		})
	}

}

type packet struct {
}

func (p *packet) Write(w *libio.Writer, b []byte) error {
	_, err := w.Write(b)
	return err
}
func (p *packet) Read(r *libio.Reader) ([]byte, error) {
	val := r.ReadBytes(6)
	return val, nil
}
